import{f as m,r as l,a as b,q as a,x as s,F as v,A as x,E as u,D as y,O as g,y as I,I as k,L as T,M as C,n as S}from"./runtime-core.esm-bundler.c752936e.js";import{s as $}from"./index.52fe7a95.js";const q=t=>(T("data-v-fa060d5c"),t=t(),C(),t),w={class:"tabs-header"},B=["onClick"],H=q(()=>u("span",{class:"tab"},null,-1)),L=[H],N=m({__name:"TabsHeader",props:{tabs:{type:Array,required:!0},activeTabIndex:{type:Number,required:!0}},emits:["update:activeTabIndex"],setup(t,{emit:_}){const p=t,n=l(),r=l(),i=e=>{e&&(r.value.style.left=`${e.offsetLeft}px`,r.value.style.width=`${e.clientWidth}px`)},f=(e,c)=>{_("update:activeTabIndex",c),S(()=>i(e.target))};return b(n,e=>{e&&setTimeout(()=>{i(n.value.children[p.activeTabIndex])},50)},{immediate:!0}),(e,c)=>(a(),s("div",w,[t.tabs?(a(),s("div",{key:0,ref_key:"tabsRef",ref:n,class:"tabs"},[(a(!0),s(v,null,x(t.tabs,({label:d},o)=>(a(),s("button",{key:`${o}${d}`,class:I([t.activeTabIndex===o?"active":"not-active"]),onClick:h=>f(h,o)},k(d),11,B))),128)),u("span",{ref_key:"highlightUnderline",ref:r,class:"highlight-underline"},L,512)],512)):y("",!0),g(e.$slots,"footer",{},void 0,!0)]))}});const D=$(N,[["__scopeId","data-v-fa060d5c"]]);export{D as default};