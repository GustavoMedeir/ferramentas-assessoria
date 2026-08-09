# -*- coding: utf-8 -*-
"""Helpers to build the Manual/Documentação HTML in the app's visual style."""
import html as _html

def esc(s):
    return _html.escape(str(s), quote=False)

CSS = """
@import url('https://fonts.googleapis.com/css2?family=Fraunces:ital,opsz,wght@0,9..144,400;0,9..144,600;0,9..144,700;1,9..144,500&family=Plus+Jakarta+Sans:ital,wght@0,400;0,500;0,600;0,700;0,800;1,400&family=JetBrains+Mono:wght@400;500;600&display=swap');

:root{
  --pine-950:#081f18; --pine-900:#0d2e26; --pine-800:#123a30;
  --gold:#c9a86a; --gold-line:#d8bd8c; --terracotta:#b5653a;
  --cream:#f2ead9; --cream-2:#efe7d6; --paper:#fffdf9;
  --ink:#1c2b26; --ink-2:#5b6b64;
  --coral:#c0503f; --coral-bg:#fbeceb; --coral-bar:#c0503f;
  --olive:#8a7332; --olive-bg:#f7f0df; --olive-bar:#c9a86a;
  --term-bg:#12201c; --term-ink:#dfe7e2; --term-comment:#7c9186;
}
*{box-sizing:border-box;}
@page{size:210mm 297mm;margin:0;}
html,body{margin:0;padding:0;}
body{font-family:"Plus Jakarta Sans",system-ui,sans-serif;color:var(--ink);background:var(--paper);font-size:13.5px;line-height:1.55;}
.page{width:210mm;min-height:297mm;box-sizing:border-box;margin:0;padding:16mm 18mm;position:relative;background:var(--paper);page-break-after:always;}
.page:last-child{page-break-after:auto;}
h1,h2,h3,h4{font-family:"Fraunces",serif;margin:0;}
p{margin:0 0 12px;}
img{max-width:100%;display:block;}

/* ---------- cover ---------- */
.cover{
  background:linear-gradient(165deg,var(--pine-800) 0%,var(--pine-900) 55%,var(--pine-950) 100%);
  color:#f4efe3;display:flex;flex-direction:column;align-items:center;justify-content:center;
  text-align:center;padding:0 80px;
}
.cover .icon{width:150px;height:150px;border-radius:34px;margin-bottom:34px;box-shadow:0 24px 60px rgba(0,0,0,.45);}
.cover h1{font-size:38px;font-weight:700;color:#f7f3e8;letter-spacing:.2px;}
.cover .subtitle{font-family:"Fraunces",serif;font-style:italic;font-size:19px;color:var(--gold);margin-top:10px;font-weight:500;}
.cover .rule{width:74px;height:2px;background:var(--gold);margin:26px auto;border:none;opacity:.8;}
.cover .meta{font-size:12.5px;color:#cfd8d1;line-height:1.9;}
.cover .meta b{color:#efe6d3;}

/* ---------- TOC ---------- */
.toc-kicker{text-align:center;font-size:11px;letter-spacing:.22em;font-weight:700;color:var(--terracotta);margin-bottom:10px;}
.toc-title{text-align:center;font-size:34px;color:var(--pine-900);letter-spacing:.04em;}
.toc-flourish{display:flex;align-items:center;justify-content:center;gap:12px;margin:14px 0 40px;}
.toc-flourish .l,.toc-flourish .r{height:1px;width:90px;background:var(--gold-line);}
.toc-flourish .d{width:6px;height:6px;background:var(--gold);transform:rotate(45deg);}
.toc-row{display:flex;gap:22px;align-items:baseline;padding:15px 0;border-bottom:1px solid var(--gold-line);}
.toc-row .n{font-family:"Fraunces",serif;font-size:30px;color:var(--gold);width:52px;flex:none;font-weight:400;}
.toc-row .t{flex:1;}
.toc-row .t .ti{font-family:"Fraunces",serif;font-weight:700;font-size:15px;color:var(--pine-900);letter-spacing:.03em;text-transform:uppercase;}
.toc-row .t .sub{font-size:12px;color:var(--terracotta);margin-top:2px;}

/* ---------- section opener ---------- */
.sec-head{display:flex;align-items:baseline;gap:18px;margin-bottom:8px;}
.sec-num{font-family:"Fraunces",serif;font-weight:400;font-size:52px;color:var(--gold);line-height:1;}
.sec-title{font-family:"Fraunces",serif;font-weight:700;font-size:25px;color:var(--pine-900);text-transform:uppercase;letter-spacing:.03em;}
.sec-rule{height:1px;background:linear-gradient(90deg,var(--gold-line),var(--cream-2));margin:14px 0 28px;}

.sub-h{font-family:"Plus Jakarta Sans",sans-serif;font-weight:800;font-size:11.5px;letter-spacing:.14em;text-transform:uppercase;color:var(--terracotta);margin:26px 0 12px;}
.sub-h:first-of-type{margin-top:0;}

/* ---------- numbered steps ---------- */
.steps{margin:14px 0;}
.step{display:flex;gap:14px;margin-bottom:12px;align-items:flex-start;}
.step .badge{flex:none;width:24px;height:24px;border:1.5px solid var(--gold);border-radius:6px;display:flex;align-items:center;justify-content:center;font-family:"Fraunces",serif;font-weight:700;font-size:13px;color:var(--gold);margin-top:1px;}
.step .txt{flex:1;padding-top:1px;}

/* ---------- callouts ---------- */
.box{background:var(--cream);border:1px solid var(--gold);border-radius:10px;margin:16px 0;overflow:hidden;}
.box .bar{background:var(--pine-900);color:#f4efe3;font-weight:800;font-size:10.5px;letter-spacing:.14em;text-transform:uppercase;padding:9px 16px;}
.box .body{padding:14px 18px;}
.box.note .body{padding-top:16px;}
.box .label{display:block;font-weight:800;font-size:10.5px;letter-spacing:.14em;text-transform:uppercase;color:var(--terracotta);margin-bottom:6px;}
.box.note{padding:0;}
.box pre{margin:0;font-family:"JetBrains Mono",monospace;font-size:12px;line-height:1.7;white-space:pre-wrap;color:var(--ink);}

.mono-chip{font-family:"JetBrains Mono",monospace;background:var(--cream-2);border:1px solid var(--gold-line);border-radius:5px;padding:1px 6px;font-size:12.5px;color:var(--pine-900);white-space:nowrap;}

/* ---------- tables ---------- */
table{width:100%;border-collapse:collapse;margin:14px 0;}
thead th{background:var(--pine-900);color:#f4efe3;text-align:left;font-size:11px;letter-spacing:.06em;text-transform:uppercase;font-weight:800;padding:10px 14px;}
tbody td{padding:10px 14px;font-size:13px;vertical-align:top;}
tbody tr:nth-child(odd){background:var(--cream-2);}
tbody tr:nth-child(even){background:#fff;}
tbody td:first-child{font-weight:700;color:var(--pine-900);}
table + .table-rule{height:1px;background:var(--gold-line);margin-top:-8px;}

/* ---------- screenshots ---------- */
.shot{border:1.5px solid var(--gold);border-radius:10px;overflow:hidden;margin:18px 0 6px;box-shadow:0 6px 20px rgba(13,46,38,.08);}
.shot img{display:block;width:100%;}
.cap{text-align:center;font-family:"Fraunces",serif;font-style:italic;color:var(--terracotta);font-size:12.5px;margin-bottom:18px;}

/* ---------- decisions/gotcha cards (doc) ---------- */
.gcard{border-radius:8px;padding:12px 16px;margin:10px 0;border-left:4px solid;}
.gcard.crit{background:var(--coral-bg);border-color:var(--coral-bar);}
.gcard.info{background:var(--olive-bg);border-color:var(--olive-bar);}
.gcard b{display:block;margin-bottom:3px;color:var(--pine-900);}

/* ---------- code blocks (doc) ---------- */
.code-light{background:#eef2ef;border-radius:9px;padding:14px 18px;margin:14px 0;font-family:"JetBrains Mono",monospace;font-size:10.6px;line-height:1.75;color:var(--ink);white-space:pre-wrap;word-break:break-word;}
.code-light .cm{color:var(--ink-2);}
.code-dark{background:var(--term-bg);border-radius:9px;padding:14px 18px;margin:14px 0;font-family:"JetBrains Mono",monospace;font-size:11.5px;line-height:1.85;color:var(--term-ink);white-space:pre-wrap;word-break:break-word;}
.code-dark .cm{color:var(--term-comment);}

/* ---------- flowchart (doc) ---------- */
.flow{display:flex;flex-wrap:wrap;gap:12px;margin:16px 0;}
.fbox{flex:1 1 200px;border:1px solid var(--gold-line);border-radius:9px;padding:12px 14px;background:#fff;}
.fbox.dark{background:var(--pine-900);color:#f4efe3;border-color:var(--pine-900);}
.fbox .ft{font-weight:800;font-size:12.5px;margin-bottom:3px;}
.fbox .fs{font-size:11.5px;color:var(--ink-2);}
.fbox.dark .fs{color:#c9d6cf;}
.farrow{display:flex;align-items:center;justify-content:center;color:var(--gold);font-size:18px;flex:0 0 20px;}

ul.plain{margin:8px 0 14px;padding-left:20px;}
ul.plain li{margin-bottom:6px;}

.closing{position:absolute;bottom:16mm;left:18mm;right:18mm;text-align:center;}
.closing .rule{height:1px;background:var(--gold-line);width:100%;margin-bottom:14px;}
.closing .txt{font-size:10.5px;letter-spacing:.18em;text-transform:uppercase;color:var(--terracotta);font-weight:700;}
.closing.plain .txt{color:var(--ink-2);letter-spacing:.04em;text-transform:none;font-weight:500;font-size:12px;}
.closing.plain .rule{display:none;}
"""

def cover(title, subtitle, tagline, version, date, author):
    return f"""
<div class="page cover">
  <img class="icon" src="images/appicon.png">
  <h1>{esc(title)}</h1>
  <div class="subtitle">{esc(subtitle)}</div>
  <hr class="rule">
  <div class="meta">{esc(tagline)}<br><b>Versão {esc(version)}</b> · {esc(date)}<br>{esc(author)}</div>
</div>"""

def toc(items):
    rows = "".join(f"""
  <div class="toc-row">
    <div class="n">{n}</div>
    <div class="t"><div class="ti">{esc(title)}</div><div class="sub">{esc(sub)}</div></div>
  </div>""" for n, title, sub in items)
    return f"""
<div class="page">
  <div class="toc-kicker">FERRAMENTAS DE ASSESSORIA</div>
  <div class="toc-title">SUMÁRIO</div>
  <div class="toc-flourish"><span class="l"></span><span class="d"></span><span class="r"></span></div>
  {rows}
</div>"""

def sec_open(num, title):
    return f"""<div class="sec-head"><div class="sec-num">{num}</div><div class="sec-title">{esc(title)}</div></div><div class="sec-rule"></div>"""

def sub(title):
    return f'<div class="sub-h">{esc(title)}</div>'

def p(text):
    return f"<p>{text}</p>"

def steps(items):
    lis = "".join(f'<div class="step"><div class="badge">{i+1}</div><div class="txt">{t}</div></div>' for i, t in enumerate(items))
    return f'<div class="steps">{lis}</div>'

def box_example(label, body_html):
    return f'<div class="box"><div class="bar">{esc(label)}</div><div class="body">{body_html}</div></div>'

def box_note(label, text):
    return f'<div class="box note"><div class="body"><span class="label">{esc(label)}</span>{text}</div></div>'

def code_example(lines):
    return "<pre>" + "\n".join(esc(l) for l in lines) + "</pre>"

def chip(s):
    return f'<span class="mono-chip">{esc(s)}</span>'

def table(headers, rows):
    th = "".join(f"<th>{esc(h)}</th>" for h in headers)
    trs = "".join("<tr>" + "".join(f"<td>{c}</td>" for c in row) + "</tr>" for row in rows)
    return f'<table><thead><tr>{th}</tr></thead><tbody>{trs}</tbody></table><div class="table-rule"></div>'

def shot(img, caption):
    return f'<div class="shot"><img src="images/{img}"></div><div class="cap">{esc(caption)}</div>'

def gcard(kind, title_, text):
    return f'<div class="gcard {kind}"><b>{esc(title_)}</b>{text}</div>'

def code_dark(lines):
    out = []
    for l in lines:
        if "#" in l:
            code, comment = l.split("#", 1)
            out.append(esc(code) + f'<span class="cm">#{esc(comment)}</span>')
        else:
            out.append(esc(l))
    return '<div class="code-dark">' + "\n".join(out) + "</div>"

def code_light(lines):
    out = []
    for l in lines:
        if "#" in l:
            code, comment = l.split("#", 1)
            out.append(esc(code) + f'<span class="cm">#{esc(comment)}</span>')
        else:
            out.append(esc(l))
    return '<div class="code-light">' + "\n".join(out) + "</div>"

def flow(boxes):
    parts = []
    for i, (t, s, dark) in enumerate(boxes):
        if i > 0:
            parts.append('<div class="farrow">&#8594;</div>')
        cls = "fbox dark" if dark else "fbox"
        parts.append(f'<div class="{cls}"><div class="ft">{esc(t)}</div><div class="fs">{esc(s)}</div></div>')
    return f'<div class="flow">{"".join(parts)}</div>'

def page(body_html, closing=None):
    close_html = ""
    if closing:
        cls = "closing plain" if closing.get("plain") else "closing"
        close_html = f'<div class="{cls}"><div class="rule"></div><div class="txt">{esc(closing["text"])}</div></div>'
    return f'<div class="page">{body_html}{close_html}</div>'

def doc(title, pages):
    return f"""<!doctype html><html lang="pt-BR"><head><meta charset="utf-8">
<title>{esc(title)}</title><style>{CSS}</style></head><body>
{''.join(pages)}
</body></html>"""
