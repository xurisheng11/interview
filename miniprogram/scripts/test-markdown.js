// Markdown 渲染器回归测试：node scripts/test-markdown.js
// 重点覆盖曾经把代码块误判成表格、导致文章详情不可读的那类输入。
const { mdToHtml } = require('../src/utils/markdown')

const sample = [
  '## 五、实战代码示例',
  '',
  '### 1. 同一 SQL 流批一体',
  '',
  '```sql',
  '-- 流模式：持续输出',
  'INSERT INTO sink',
  'SELECT user_id, COUNT(*) AS cnt',
  'FROM clicks',
  'GROUP BY user_id;',
  '',
  "SET 'execution.runtime-mode' = 'BATCH';",
  '```',
  '',
  '### 3. 架构简图',
  '',
  '```',
  '┌──────────────────────────────┐',
  '│   统一 API 层 (DataStream / Table)   │',
  '├──────────────────────────────┤',
  '| 流: Pipelined Shuffle 批: Blocking Shuffle |',
  '└──────────────────────────────┘',
  '```',
  '',
  '正文里的 **加粗**、`inline()` 代码、[链接](https://a.b) 和表格：',
  '',
  '| 维度 | 流 | 批 |',
  '|---|---|---|',
  '| 延迟 | 毫秒 | 分钟 |',
  '',
  '> 引用一句话',
  '',
  '- 列表项 A',
  '- 列表项 B',
  '',
  '结尾段落：<script>alert(1)</script> 应该被转义。'
].join('\n')

const html = mdToHtml(sample)
// 代码块里的空格被转成字面量 NBSP 以保对齐，断言前先归一化回来
const norm = html.replace(/\u00a0/g, ' ')

function assert(name, cond) {
  console.log((cond ? '✅' : '❌') + ' ' + name)
  if (!cond) process.exitCode = 1
}

assert('不残留 ``` 围栏标记', norm.indexOf('```') < 0)
assert('SQL 注释进入代码块', norm.indexOf('-- 流模式：持续输出') >= 0)
assert('代码块是深色卡片', html.indexOf('background:#1e2530') >= 0)
assert('带语言标签', html.indexOf('>sql<') >= 0)
assert('ASCII 框图未被误判成表格（全文只应有 1 个真表格）', norm.indexOf('统一 API 层') >= 0 && norm.split('<table').length === 2)
assert('代码块用 NBSP 保对齐且不依赖实体解码', html.indexOf('\u00a0') >= 0 && html.indexOf('&nbsp;') < 0)
assert('框图用 <br/> 保行结构', html.indexOf('<br/>') >= 0)
assert('框图字符未被当列表符号吃掉', norm.indexOf('└──') >= 0)
assert('表格仍正常渲染', html.indexOf('<th style=') >= 0 && norm.indexOf('延迟') >= 0)
assert('加粗生效', html.indexOf('<strong') >= 0)
assert('行内代码生效', norm.indexOf('inline()') >= 0)
assert('链接渲染成蓝文字且无残留中括号', html.indexOf('https://a.b') < 0 && html.indexOf('[链接]') < 0 && html.indexOf('>链接<') >= 0)
assert('引用块生效', html.indexOf('<blockquote') >= 0)
assert('列表生效', html.indexOf('<li style=') >= 0)
assert('HTML 注入被转义', html.indexOf('<script>') < 0 && html.indexOf('&lt;script&gt;') >= 0)
assert('## → h3、### → h4（不占 h1/h2）', html.indexOf('<h3 style=') >= 0 && html.indexOf('<h4 style=') >= 0)
assert('代码块内不跑行内 markdown（* 不被吃掉）', norm.indexOf('COUNT(*)') >= 0)

// 未闭合围栏：按 CommonMark 语义吃到文末，但不能丢内容、不能死循环
const broken = mdToHtml('```\ncode line\n后面还有正文').replace(/\u00a0/g, ' ')
assert('未闭合围栏不丢内容', broken.indexOf('code line') >= 0 && broken.indexOf('后面还有正文') >= 0)

console.log('\n--- 片段预览 ---')
console.log(html.substring(0, 420).replace(/></g, '>\n<'))
