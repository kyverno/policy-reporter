import G from"./DocsAside.62934c7c.js";import J from"./ProseCodeInline.7d97da57.js";import R from"./Alert.50a3c16b.js";import U from"./DocsPageBottom.54ae9808.js";import K from"./DocsPrevNext.1e059232.js";import{_ as Q}from"./DocsAsideTree.e4337552.js";import W from"./DocsToc.15b9f88b.js";import{e as X,c as Y,u as Z,f as oo,d as eo}from"./Container.68928775.js";import{o as to}from"./index.550f448e.js";import{f as no,h as d,r as I,N as so,o as ao,q as i,B as $,C as g,u as t,x,G as c,D as y,E as p,O as co,J as k,I as ro,F as lo,y as f,L as io,M as po}from"./runtime-core.esm-bundler.c752936e.js";/* empty css                           *//* empty css                      */import"./cookie.722ca6ef.js";import"./ContentSlot.89901889.js";/* empty css                  */import"./ProseA.2fc38c83.js";/* empty css                   */import"./EditOnLink.vue.a3dcfb8b.js";/* empty css                           *//* empty css                         */import"./DocsTocLinks.8ad87046.js";/* empty css                         *//* empty css                    */const uo=u=>(io("data-v-64504251"),u=u(),po(),u),_o={key:1,class:"toc"},mo={class:"toc-wrapper"},fo=uo(()=>p("span",{class:"title"},"Table of Contents",-1)),vo=no({__name:"DocsPageLayout",setup(u){const{page:s,navigation:V}=X(),{config:v}=Y(),A=to(),M=(o,e=!0)=>{var n;return typeof((n=s.value)==null?void 0:n[o])<"u"?s.value[o]:e},T=d(()=>{var o,e,n;return!s.value||((n=(e=(o=s.value)==null?void 0:o.body)==null?void 0:e.children)==null?void 0:n.length)>0}),C=d(()=>{var o,e,n,_,m;return((o=s.value)==null?void 0:o.toc)!==!1&&((m=(_=(n=(e=s.value)==null?void 0:e.body)==null?void 0:n.toc)==null?void 0:_.links)==null?void 0:m.length)>=2}),E=d(()=>{var o,e;return((o=s.value)==null?void 0:o.aside)!==!1&&((e=V.value)==null?void 0:e.length)>0}),F=d(()=>M("bottom",!0)),r=I(!1),a=I(null),h=()=>A.path.split("/").slice(0,2).join("/"),l=Z("asideScroll",()=>{var o;return{parentPath:h(),scrollTop:((o=a.value)==null?void 0:o.scrollTop)||0}});function S(){a.value&&(a.value.scrollHeight===0&&setTimeout(S,0),a.value.scrollTop=l.value.scrollTop)}return so(()=>{l.value.parentPath!==h()?(l.value.parentPath=h(),l.value.scrollTop=0):S()}),ao(()=>{a.value&&(l.value.scrollTop=a.value.scrollTop)}),(o,e)=>{var P,b,B,D,N,w;const n=G,_=J,m=R,H=U,L=K,O=Q,j=W,q=oo;return i(),$(q,{fluid:(b=(P=t(v))==null?void 0:P.main)==null?void 0:b.fluid,padded:(D=(B=t(v))==null?void 0:B.main)==null?void 0:D.padded,class:f(["docs-page-content",{fluid:(w=(N=t(v))==null?void 0:N.main)==null?void 0:w.fluid}])},{default:g(()=>[t(E)?(i(),x("aside",{key:0,ref_key:"asideNav",ref:a,class:"aside-nav"},[c(n,{class:"app-aside"})],512)):y("",!0),p("article",{class:f(["page-body",{"with-toc":t(C)}])},[t(T)?co(o.$slots,"default",{key:0},void 0,!0):(i(),$(m,{key:1,type:"info"},{default:g(()=>[k(" Start writing in "),c(_,null,{default:g(()=>[k("content/"+ro(t(s)._file),1)]),_:1}),k(" to see this page taking shape. ")]),_:1})),t(T)&&t(s)&&t(F)?(i(),x(lo,{key:2},[c(H),c(L)],64)):y("",!0)],2),t(C)?(i(),x("div",_o,[p("div",mo,[p("button",{onClick:e[0]||(e[0]=z=>r.value=!t(r))},[fo,c(O,{name:"heroicons-outline:chevron-right",class:f(["icon",[t(r)&&"rotate"]])},null,8,["class"])]),p("div",{class:f(["docs-toc-wrapper",[t(r)&&"opened"]])},[c(j,{onMove:e[1]||(e[1]=z=>r.value=!1)})],2)])])):y("",!0)]),_:3},8,["fluid","padded","class"])}}}),Oo=eo(vo,[["__scopeId","data-v-64504251"]]);export{Oo as default};