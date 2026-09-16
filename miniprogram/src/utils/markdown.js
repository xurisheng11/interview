// 轻量 Markdown → HTML 转换器（输出供 <rich-text> 使用）
// 支持：标题 #~######、围栏代码块 ```lang、加粗 **x**、斜体 *x*、行内代码 `x`、链接、
// 无序/有序列表、表格、引用块、分隔线、段落。样式全部用内联 px（rich-text 不支持 rpx）。

function escapeHtml(s) {
  return s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')
}

function inlineMd(s) {
  s = s.replace(/`([^`]+)`/g, '<code style="background:#f2f4f7;color:#c7254e;padding:1px 5px;border-radius:4px;font-size:13px;">$1</code>')
  // 图片在正文里无法加载时至少保留描述文字，链接则渲染成可辨识的蓝色文字（rich-text 内不可点）
  s = s.replace(/!\[([^\]]*)\]\(([^)\s]+)[^)]*\)/g, '<span style="color:#8b98a9;font-size:13px;">🖼 $1</span>')
  s = s.replace(/\[([^\]]+)\]\(([^)\s]+)[^)]*\)/g, '<span style="color:#1677ff;text-decoration:underline;">$1</span>')
  s = s.replace(/\*\*([^*]+)\*\*/g, '<strong style="color:#1a1a2e;">$1</strong>')
  s = s.replace(/(^|[^*])\*([^*\s][^*]*)\*/g, '$1<em>$2</em>')
  return s
}

var STYLES = {
  h1: 'font-size:21px;font-weight:800;color:#1a1a2e;margin:20px 0 12px;line-height:1.5;',
  h2: 'font-size:18px;font-weight:800;color:#1a1a2e;margin:20px 0 10px;padding-left:9px;border-left:4px solid #1677ff;line-height:1.5;',
  h3: 'font-size:16px;font-weight:700;color:#1a1a2e;margin:16px 0 8px;line-height:1.5;',
  h4: 'font-size:15px;font-weight:700;color:#2e5c9b;margin:14px 0 6px;line-height:1.5;',
  p: 'font-size:15px;color:#333a45;line-height:1.85;margin:10px 0;word-break:break-word;',
  ul: 'margin:8px 0;padding-left:22px;',
  ol: 'margin:8px 0;padding-left:22px;',
  li: 'font-size:15px;color:#333a45;line-height:1.8;margin:4px 0;word-break:break-all;',
  quote: 'margin:12px 0;padding:10px 12px;background:#f6f9ff;border-left:4px solid #1677ff;color:#5c6470;font-size:14px;line-height:1.8;border-radius:0 6px 6px 0;',
  table: 'border-collapse:collapse;width:100%;margin:14px 0;font-size:13px;',
  th: 'border:1px solid #e3e8ef;background:#f0f5ff;color:#1a1a2e;font-weight:700;padding:7px 8px;text-align:left;line-height:1.6;word-break:break-all;',
  td: 'border:1px solid #e8ecf2;color:#333a45;padding:7px 8px;line-height:1.6;word-break:break-all;',
  hr: 'border:none;border-top:1px solid #eceff4;margin:18px 0;'
}

// 代码块：深色卡片 + 等宽字体。
// 空格统一换成字面量 NBSP（U+00A0）：它不会被 HTML 归一化吐掉，
// 又不依赖 rich-text 对 &nbsp; 实体的解码，ASCII 框图的对齐才能保住。
var NBSP = String.fromCharCode(0xa0)
var CODE_WRAP = 'margin:14px 0;border-radius:10px;overflow:hidden;background:#1e2530;'
var CODE_HEAD = 'padding:5px 12px;background:#2b3542;color:#8b98a9;font-size:11px;letter-spacing:1px;font-family:Menlo,Consolas,monospace;'
var CODE_BODY = 'padding:11px 12px;color:#dfe7f1;font-size:12px;line-height:1.75;white-space:pre-wrap;word-break:break-word;font-family:Menlo,Consolas,Monaco,"Courier New",monospace;'

function renderCodeBlock(lang, lines) {
  var body = lines.map(function (ln) {
    return escapeHtml(ln).replace(/\t/g, '    ').replace(/ /g, NBSP)
  }).join('<br/>')
  if (!body) body = NBSP
  var head = lang ? '<div style="' + CODE_HEAD + '">' + escapeHtml(lang) + '</div>' : ''
  return '<div style="' + CODE_WRAP + '">' + head + '<div style="' + CODE_BODY + '">' + body + '</div></div>'
}

function isTableRow(line) {
  return line.indexOf('|') === 0 || (line.indexOf('|') > 0 && line.split('|').length >= 3 && /\|/.test(line))
}

function splitRow(line) {
  var s = line.trim()
  if (s.charAt(0) === '|') s = s.substring(1)
  if (s.charAt(s.length - 1) === '|') s = s.substring(0, s.length - 1)
  var cells = s.split('|')
  for (var i = 0; i < cells.length; i++) cells[i] = cells[i].trim()
  return cells
}

function isSeparatorRow(line) {
  return /^\|?[\s:|-]+\|?$/.test(line.trim()) && line.indexOf('-') >= 0
}

/**
 * markdown 文本 → HTML 字符串（rich-text nodes 可直接使用）
 */
function mdToHtml(md) {
  if (!md) return ''
  // 兼容被转义的换行符（AI 生成内容常见 "\n" 字面量）
  var NL = String.fromCharCode(10)
  var text = String(md).split('\\r\\n').join(NL).split('\\n').join(NL)
  var lines = text.split(NL)
  var html = []
  var listType = null // 'ul' | 'ol'
  var i = 0

  function closeList() {
    if (listType) {
      html.push(listType === 'ul' ? '</ul>' : '</ol>')
      listType = null
    }
  }

  while (i < lines.length) {
    var raw = lines[i].replace(/\s+$/, '')
    var line = raw.trim()

    // 空行：结束列表
    if (line === '') { closeList(); i++; continue }

    // 围栏代码块：必须排在表格判断之前，
    // 否则代码里的 | 和 +----+ 会被当成表格，渲染出贯穿屏幕的表格边框
    var fence = line.match(/^`{3,}\s*([A-Za-z0-9_+\-#.]*?)\s*`*$/)
    if (fence) {
      closeList()
      var lang = fence[1].toLowerCase()
      var codeLines = []
      i++
      while (i < lines.length) {
        var closing = lines[i].trim()
        if (/^`{3,}$/.test(closing)) { i++; break }
        codeLines.push(lines[i].replace(/\s+$/, ''))
        i++
      }
      html.push(renderCodeBlock(lang, codeLines))
      continue
    }

    // 表格块：连续的 | 行
    if (line.indexOf('|') >= 0 && (line.charAt(0) === '|' || /\|\s/.test(line)) && line.split('|').length >= 3) {
      closeList()
      var rows = []
      while (i < lines.length) {
        var t = lines[i].trim()
        if (t === '' || (t.indexOf('|') < 0 && t.split('|').length < 3)) break
        if (!isSeparatorRow(t)) rows.push(splitRow(t))
        i++
      }
      html.push('<table style="' + STYLES.table + '">')
      for (var r = 0; r < rows.length; r++) {
        html.push('<tr>')
        for (var c = 0; c < rows[r].length; c++) {
          var cell = inlineMd(escapeHtml(rows[r][c]))
          if (r === 0) html.push('<th style="' + STYLES.th + '">' + cell + '</th>')
          else html.push('<td style="' + STYLES.td + '">' + cell + '</td>')
        }
        html.push('</tr>')
      }
      html.push('</table>')
      continue
    }

    // 标题（#~###### 统一映射到 h2~h4，页面已有文章大标题所以不占 h1）
    var h = line.match(/^(#{1,6})\s*(.+)$/)
    if (h) {
      closeList()
      var level = Math.min(h[1].length + 1, 4) // # → h2（页面已有大标题）
      html.push('<h' + level + ' style="' + STYLES['h' + level] + '">' + inlineMd(escapeHtml(h[2])) + '</h' + level + '>')
      i++
      continue
    }

    // 分隔线
    if (/^(-{3,}|\*{3,}|_{3,})$/.test(line)) {
      closeList()
      html.push('<hr style="' + STYLES.hr + '"/>')
      i++
      continue
    }

    // 引用
    var q = line.match(/^>\s*(.+)$/)
    if (q) {
      closeList()
      html.push('<blockquote style="' + STYLES.quote + '">' + inlineMd(escapeHtml(q[1])) + '</blockquote>')
      i++
      continue
    }

    // 无序列表
    var ul = line.match(/^[-*•]\s+(.+)$/)
    if (ul) {
      if (listType !== 'ul') { closeList(); html.push('<ul style="' + STYLES.ul + '">'); listType = 'ul' }
      html.push('<li style="' + STYLES.li + '">' + inlineMd(escapeHtml(ul[1])) + '</li>')
      i++
      continue
    }

    // 有序列表
    var ol = line.match(/^\d{1,2}[.、)]\s*(.+)$/)
    if (ol) {
      if (listType !== 'ol') { closeList(); html.push('<ol style="' + STYLES.ol + '">'); listType = 'ol' }
      html.push('<li style="' + STYLES.li + '">' + inlineMd(escapeHtml(ol[1])) + '</li>')
      i++
      continue
    }

    // 普通段落
    closeList()
    html.push('<p style="' + STYLES.p + '">' + inlineMd(escapeHtml(line)) + '</p>')
    i++
  }

  closeList()
  return html.join('')
}

module.exports = { mdToHtml: mdToHtml }
