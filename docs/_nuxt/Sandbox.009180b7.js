import _ from"./TabsHeader.ca4eb532.js";import{u as x}from"./entry.44fa159c.js";import{f as h,r,h as g,N as S,q as n,x as l,B as k,u as c,D as y}from"./runtime-core.esm-bundler.c752936e.js";/* empty css                    */import{d as $}from"./Container.68928775.js";import"./index.550f448e.js";import"./DocsAsideTree.e4337552.js";import"./cookie.722ca6ef.js";import"./query.c3f7607a.js";const w={class:"sandbox"},B=["src"],C={key:2},I=h({__name:"Sandbox",props:{src:{type:String,default:""},repo:{type:String,default:""},branch:{type:String,default:""},dir:{type:String,default:""},file:{type:String,default:"app.vue"}},setup(i){const t=i,p=x(),s={CodeSandBox:()=>`https://codesandbox.io/embed/github/${t.repo}/tree/${t.branch}/${t.dir}?hidenavigation=1&theme=${p.value}`,StackBlitz:()=>`https://stackblitz.com/github/${t.repo}/tree/${t.branch}/${t.dir}?embed=1&file=${t.file}&theme=${p.value}`},u=Object.keys(s).map(e=>({label:e})),d=r(-1),m=r(),o=r(""),a=r(""),b=e=>{a.value=e,o.value=t.src||s[a.value](),localStorage.setItem("docus_sandbox",e)};g(()=>{var e;return(e=o.value)==null?void 0:e.replace("?embed=1&","?").replace("/embed/","/s/")});const f=e=>{d.value=e,b(u[e].label)};return S(()=>{a.value=window.localStorage.getItem("docus_sandbox")||"CodeSandBox",o.value=t.src||s[a.value](),d.value=Object.keys(s).indexOf(a.value)}),(e,T)=>{const v=_;return n(),l("div",w,[i.src?y("",!0):(n(),k(v,{key:0,ref_key:"tabs",ref:m,"active-tab-index":c(d),tabs:c(u),"onUpdate:activeTabIndex":f},null,8,["active-tab-index","tabs"])),c(o)?(n(),l("iframe",{key:1,src:c(o),title:"Sandbox editor",sandbox:"allow-modals allow-forms allow-popups allow-scripts allow-same-origin"},null,8,B)):(n(),l("span",C,"Loading Sandbox..."))])}}}),L=$(I,[["__scopeId","data-v-90b7615f"]]);export{L as default};