package send_test

import (
	"bufio"
	"encoding/base64"
	"encoding/csv"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net"
	"net/mail"
	"net/textproto"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type smtpCapture struct {
	listener net.Listener
	messages chan []byte
	wg       sync.WaitGroup
}

func newSMTPCapture(t *testing.T, reject bool) *smtpCapture {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	s := &smtpCapture{listener: listener, messages: make(chan []byte, 20)}
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			s.wg.Add(1)
			go func() {
				defer s.wg.Done()
				defer conn.Close()
				_ = conn.SetDeadline(time.Now().Add(10 * time.Second))
				reader := textproto.NewReader(bufio.NewReader(conn))
				fmt.Fprint(conn, "220 localhost ESMTP\r\n")
				for {
					line, err := reader.ReadLine()
					if err != nil {
						return
					}
					switch {
					case strings.HasPrefix(line, "EHLO"), strings.HasPrefix(line, "HELO"):
						fmt.Fprint(conn, "250 localhost\r\n")
					case strings.HasPrefix(line, "MAIL FROM"), strings.HasPrefix(line, "RCPT TO"), line == "RSET":
						fmt.Fprint(conn, "250 OK\r\n")
					case line == "DATA":
						fmt.Fprint(conn, "354 Send message\r\n")
						data, err := io.ReadAll(reader.DotReader())
						if err != nil {
							return
						}
						s.messages <- data
						if reject {
							fmt.Fprint(conn, "552 Message size exceeds fixed maximum message size\r\n")
						} else {
							fmt.Fprint(conn, "250 queued\r\n")
						}
					case line == "QUIT":
						fmt.Fprint(conn, "221 bye\r\n")
						return
					default:
						fmt.Fprint(conn, "500 unsupported\r\n")
					}
				}
			}()
		}
	}()
	t.Cleanup(func() { listener.Close(); s.wg.Wait() })
	return s
}

func (s *smtpCapture) config() string {
	addr := s.listener.Addr().(*net.TCPAddr)
	return fmt.Sprintf("  smtp:\n    host: 127.0.0.1\n    port: %d\n    from: reports@example.test\n", addr.Port)
}

type decodedMail struct {
	to, subject, body, filename, contentType string
	attachment                               []byte
}

func decodeMail(t *testing.T, raw []byte) decodedMail {
	t.Helper()
	m, err := mail.ReadMessage(strings.NewReader(string(raw)))
	require.NoError(t, err)
	result := decodedMail{to: m.Header.Get("To"), subject: m.Header.Get("Subject")}
	var walk func(textproto.MIMEHeader, io.Reader)
	walk = func(header textproto.MIMEHeader, body io.Reader) {
		media, params, err := mime.ParseMediaType(header.Get("Content-Type"))
		require.NoError(t, err)
		if strings.HasPrefix(media, "multipart/") {
			reader := multipart.NewReader(body, params["boundary"])
			for {
				part, err := reader.NextPart()
				if err == io.EOF {
					break
				}
				require.NoError(t, err)
				walk(part.Header, part)
			}
			return
		}
		switch strings.ToLower(header.Get("Content-Transfer-Encoding")) {
		case "base64":
			body = base64.NewDecoder(base64.StdEncoding, body)
		case "quoted-printable":
			body = quotedprintable.NewReader(body)
		}
		data, err := io.ReadAll(body)
		require.NoError(t, err)
		if header.Get("Content-Disposition") != "" {
			disposition, params, err := mime.ParseMediaType(header.Get("Content-Disposition"))
			require.NoError(t, err)
			require.Equal(t, "attachment", disposition)
			result.filename = params["filename"]
			result.contentType = media
			result.attachment = data
		} else {
			result.body += string(data)
		}
	}
	walk(textproto.MIMEHeader(m.Header), m.Body)
	return result
}

func TestCSVCommandSMTP(t *testing.T) {
	for _, reportType := range []string{"summary", "violations"} {
		for _, format := range []string{"", "csv"} {
			t.Run(reportType+"/"+format, func(t *testing.T) {
				api := localReportsAPI(t, false)
				defer api.Close()
				smtp := newSMTPCapture(t, false)
				logs := runReportCommand(t, api.URL, reportType, smtp.config(), fmt.Sprintf("    to: [reader@example.test]\n    attachmentFormat: %q\n", format))
				require.Contains(t, logs, "email sent to")
				m := decodeMail(t, <-smtp.messages)
				require.Contains(t, m.to, "reader@example.test")
				require.Contains(t, m.subject, "("+reportType+") on local")
				if format == "" {
					require.Empty(t, m.attachment)
					require.Contains(t, m.body, "team-a")
					return
				}
				require.Equal(t, reportType+".csv", m.filename)
				require.Equal(t, "text/csv", m.contentType)
				require.Contains(t, m.body, "attached")
				require.NotContains(t, m.body, "team-a")
				rows, err := csv.NewReader(strings.NewReader(string(m.attachment))).ReadAll()
				require.NoError(t, err)
				require.Len(t, rows, 5)
				if reportType == "summary" {
					require.Equal(t, []string{"local", "kyverno", "namespace", "team-a", "5", "2", "0", "0", "0"}, rows[2])
				} else {
					require.Equal(t, []string{"local", "kyverno", "namespace", "team-a", "require-label", "app-label", "Deployment", "a1", "fail", "medium", "Missing label, \"app\"\n雪"}, rows[2])
				}
			})
		}
	}
}

func TestCSVCommandMixedChannels(t *testing.T) {
	for _, reportType := range []string{"summary", "violations"} {
		t.Run(reportType, func(t *testing.T) {
			api := localReportsAPI(t, false)
			defer api.Close()
			smtp := newSMTPCapture(t, false)
			runReportCommand(t, api.URL, reportType, smtp.config(), `    to: [parent@example.test]
    attachmentFormat: csv
    channels:
    - to: [a@example.test]
      attachmentFormat: csv
      filter:
        disableClusterReports: true
        namespaces:
          include: [team-a]
        sources:
          include: [kyverno]
    - to: [b@example.test]
      filter:
        disableClusterReports: true
        namespaces:
          include: [team-b]
    - to: [empty@example.test]
      attachmentFormat: csv
      filter:
        disableClusterReports: true
        namespaces:
          include: [missing]
    - attachmentFormat: csv
`)
			require.Len(t, smtp.messages, 3)
			for range 3 {
				m := decodeMail(t, <-smtp.messages)
				switch {
				case strings.Contains(m.to, "a@example.test"):
					rows, err := csv.NewReader(strings.NewReader(string(m.attachment))).ReadAll()
					require.NoError(t, err)
					for _, row := range rows[1:] {
						require.Equal(t, "kyverno", row[1])
						require.Equal(t, "namespace", row[2])
						require.Equal(t, "team-a", row[3])
					}
				case strings.Contains(m.to, "b@example.test"):
					require.Empty(t, m.attachment)
					require.Contains(t, m.body, "team-b")
					require.NotContains(t, m.body, "team-a")
				case strings.Contains(m.to, "parent@example.test"):
					require.NotEmpty(t, m.attachment)
				default:
					t.Fatalf("unexpected recipient %s", m.to)
				}
			}
		})
	}
}

func TestCSVCommandEmptyAndRejected(t *testing.T) {
	for _, reportType := range []string{"summary", "violations"} {
		for _, format := range []string{"", "csv"} {
			t.Run(reportType+"/empty/"+format, func(t *testing.T) {
				api := localReportsAPI(t, true)
				defer api.Close()
				smtp := newSMTPCapture(t, false)
				logs := runReportCommand(t, api.URL, reportType, smtp.config(), fmt.Sprintf("    to: [reader@example.test]\n    attachmentFormat: %q\n    channels:\n    - to: [channel@example.test]\n      attachmentFormat: csv\n", format))
				require.Contains(t, logs, "skip email - no results to send")
				require.Len(t, smtp.messages, 1)
				m := decodeMail(t, <-smtp.messages)
				if format == "csv" {
					rows, err := csv.NewReader(strings.NewReader(string(m.attachment))).ReadAll()
					require.NoError(t, err)
					require.Len(t, rows, 1)
				} else {
					require.Empty(t, m.attachment)
				}
			})
		}
		t.Run(reportType+"/rejected", func(t *testing.T) {
			api := localReportsAPI(t, false)
			defer api.Close()
			smtp := newSMTPCapture(t, true)
			logs := runReportCommand(t, api.URL, reportType, smtp.config(), "    to: [reader@example.test]\n    attachmentFormat: csv\n")
			require.Contains(t, logs, "failed to send report")
			require.Contains(t, logs, "552")
			require.NotContains(t, logs, "email sent to")
		})
	}
}

func TestCSVCommandNoRecipients(t *testing.T) {
	api := localReportsAPI(t, false)
	defer api.Close()
	for _, reportType := range []string{"summary", "violations"} {
		t.Run(reportType, func(t *testing.T) {
			smtp := newSMTPCapture(t, false)
			logs := runReportCommand(t, api.URL, reportType, smtp.config(), "    attachmentFormat: csv\n")
			require.Contains(t, logs, "skipped - no email configured")
			require.Empty(t, smtp.messages)
			logs = runReportCommand(t, api.URL, reportType, smtp.config(), `    attachmentFormat: csv
    filter:
      disableClusterReports: true
      sources:
        include: [kyverno]
    channels:
    - to: [channel@example.test]
      attachmentFormat: csv
      filter:
        disableClusterReports: true
    - to: [excluded@example.test]
      filter:
        disableClusterReports: true
        sources:
          include: [trivy]
`)
			require.Contains(t, logs, "skipped - no email configured")
			require.Len(t, smtp.messages, 1)
			m := decodeMail(t, <-smtp.messages)
			require.Contains(t, m.to, "channel@example.test")
			rows, err := csv.NewReader(strings.NewReader(string(m.attachment))).ReadAll()
			require.NoError(t, err)
			for _, row := range rows[1:] {
				require.Equal(t, "kyverno", row[1])
				require.Equal(t, "namespace", row[2])
			}
		})
	}
}
