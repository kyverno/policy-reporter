# Email report CSV attachments

Summary and violation emails can optionally attach CSV data instead of putting
all report details in the email body. CSV works with both SMTP and Microsoft
Graph. This feature does not produce native XLSX files.

## Configuration

Set `attachmentFormat: csv` independently for each report type or channel:

```yaml
emailReports:
  clusterName: production
  summary:
    to: [operations@example.org]
    attachmentFormat: csv
    channels:
      - to: [team-a@example.org]
        attachmentFormat: csv
        filter:
          disableClusterReports: true
          namespaces:
            include: [team-a]
      - to: [html-reader@example.org]
        # No attachmentFormat: retains the existing detailed email.
  violations:
    to: [security@example.org]
    attachmentFormat: csv
```

Omitting the setting or using `attachmentFormat: ""` preserves the existing
email and custom templates, including the embedded-template fallback. Other
values fail configuration loading. The existing `format` setting still controls
the email body's content type; it is not an export format.

Channels do not inherit attachment mode. Their filters operate on the data
already selected by the report's top-level filter; a channel cannot recover data
excluded there. HTML API responses remain detailed HTML regardless of email
attachment settings.

The same settings are supported in Helm values. Set
`emailReports.summary.enabled: true` and/or
`emailReports.violations.enabled: true` and configure the existing SMTP or Graph
transport to schedule delivery.

## CSV contents

`summary.csv` columns:

```csv
Cluster,Source,Scope,Namespace,Pass,Fail,Warn,Error,Skip
production,kyverno,namespace,team-a,8,2,1,0,0
production,kyverno,cluster,,4,1,0,0,0
```

Each source has one row per namespace, using the existing report-summary counts.
When cluster reports are enabled, each source also has a cluster row, including
zero counts where the current email displays an empty cluster summary. There
are no extra total rows. Scope is `cluster` or `namespace`; a cluster row has an
empty namespace, not a synthetic namespace name.

`violations.csv` columns:

```csv
Cluster,Source,Scope,Namespace,Policy,Rule,Kind,Name,Status,Severity,Message
production,kyverno,namespace,team-a,require-label,app-label,Deployment,nginx,fail,medium,Missing app label
```

Rows follow the existing warn/fail/error selection and resource expansion: one
row per subject, report scope as fallback, or an empty kind/name when no resource
is available. Namespace identifies the report's namespace, as in the email
headings. Message and severity retain available report values; missing values
remain empty. The existing rule-to-description fallback is retained.

Sources and namespaces are sorted; violation statuses follow the email order
(warn, fail, error), with rows sorted by policy, rule, kind, name, status,
severity, and message. CSV is UTF-8 without a BOM, with standard CSV quoting for
commas, quotes and multiline text. When opening in spreadsheet software, select
UTF-8 if it is not detected automatically.

To reduce spreadsheet formula injection, text cells are prefixed with an
apostrophe if they begin with tab, CR, or LF, or their first character after
leading Unicode whitespace, control, or format characters is `=`, `+`, `-`, or
`@`. The original text follows the apostrophe unchanged. This can make the
apostrophe visible in CSV readers. Counts remain numeric. Source report data and
HTML output are not modified.

## Empty reports and failures

Existing send decisions are unchanged: configured top-level recipients can
receive an empty report (a header-only CSV); channels skip when their filtered
source list is empty. A source with no violations can still produce a
header-only violation CSV. No-recipient groups are skipped.

CSV mode sends a short explanatory body and one attachment. A generation or
provider error is logged without sending a substitute email that omits details.
Attachments are held in memory and are not split or compressed. There is no
universal size limit: SMTP/Graph rejection is returned by the sender and logged
by the command, without a success log for that failed send. The existing command
exit behavior is unchanged: individual delivery failures are logged but do not
necessarily produce a nonzero command exit status.

Microsoft Graph uses an inline JSON `fileAttachment` with base64 content, not an
upload session. Large attachments may be rejected; consult Microsoft's
[large attachment documentation](https://learn.microsoft.com/en-us/graph/outlook-large-attachments).
A Graph `202 Accepted` response establishes acceptance, not final delivery.
