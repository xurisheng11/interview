<template>
  <div class="markdown-render" v-html="renderedHtml"></div>
</template>

<script>
import { marked } from 'marked'

export default {
  name: 'MarkdownRender',
  props: {
    content: {
      type: String,
      default: ''
    }
  },
  computed: {
    renderedHtml() {
      if (!this.content) return ''
      try {
        return marked.parse(this.content)
      } catch (e) {
        return this.content
      }
    }
  }
}
</script>

<style scoped>
.markdown-render {
  font-size: 14px;
  line-height: 1.8;
  color: #333;
  word-break: break-word;
}
.markdown-render :deep(h1),
.markdown-render :deep(h2),
.markdown-render :deep(h3) {
  margin: 12px 0 6px;
  font-weight: bold;
  color: #222;
}
.markdown-render :deep(p) {
  margin: 6px 0;
}
.markdown-render :deep(ul),
.markdown-render :deep(ol) {
  padding-left: 20px;
  margin: 6px 0;
}
.markdown-render :deep(li) {
  margin-bottom: 3px;
}
.markdown-render :deep(code) {
  background: #f5f5f5;
  border-radius: 3px;
  padding: 1px 5px;
  font-family: 'Courier New', monospace;
  font-size: 13px;
  color: #c7254e;
}
.markdown-render :deep(pre) {
  background: #282c34;
  color: #abb2bf;
  padding: 12px 16px;
  border-radius: 6px;
  overflow-x: auto;
  margin: 8px 0;
}
.markdown-render :deep(pre code) {
  background: transparent;
  color: inherit;
  padding: 0;
  font-size: 13px;
}
.markdown-render :deep(blockquote) {
  border-left: 4px solid #ff9900;
  margin: 8px 0;
  padding: 4px 12px;
  color: #666;
  background: #fffbf0;
}
.markdown-render :deep(table) {
  border-collapse: collapse;
  width: 100%;
  margin: 8px 0;
}
.markdown-render :deep(th),
.markdown-render :deep(td) {
  border: 1px solid #e0e0e0;
  padding: 6px 12px;
  text-align: left;
}
.markdown-render :deep(th) {
  background: #f5f5f5;
  font-weight: bold;
}
</style>
