<template>
  <div class="help-page">
    <!-- 顶部标题 -->
    <div class="help-header">
      <div class="header-content">
        <h1 class="header-title">📖 MockInterview 使用指南</h1>
        <p class="header-sub">欢迎使用 MockInterview！这里是您快速上手的最佳入口</p>
      </div>
    </div>

    <!-- 核心功能区 -->
    <div class="help-section">
      <h2 class="section-title">🎯 核心功能</h2>
      <div class="feature-grid">
        <div class="feature-card" @click="$router.push('/interview/config')">
          <div class="feature-icon">🚀</div>
          <div class="feature-info">
            <h3>AI 模拟面试</h3>
            <p>上传简历，选择岗位和难度，AI 面试官将为您生成个性化的面试场景，支持多轮面试。</p>
          </div>
        </div>
        
        <div class="feature-card" @click="$router.push('/questions')">
          <div class="feature-icon">📝</div>
          <div class="feature-info">
            <h3>题库练习</h3>
            <p>海量真实面试题目，涵盖各公司和岗位，随时随地练习，提升答题技巧。</p>
          </div>
        </div>
        
        <div class="feature-card" @click="$router.push('/company/intel')">
          <div class="feature-icon">🏢</div>
          <div class="feature-info">
            <h3>公司面试知识库</h3>
            <p>了解各公司的面试流程、常见问题、薪资范围，提前做好准备。</p>
          </div>
        </div>
        
        <div class="feature-card" @click="$router.push('/community')">
          <div class="feature-icon">🌐</div>
          <div class="feature-info">
            <h3>知识社区</h3>
            <p>与求职者交流经验，分享面试技巧，获取第一手求职信息。</p>
          </div>
        </div>
      </div>
    </div>

    <!-- 快速开始 -->
    <div class="help-section">
      <h2 class="section-title">⚡ 快速开始</h2>
      <el-steps :active="currentStep" finish-status="success" align-center class="start-steps">
        <el-step title="上传简历" description="在简历管理中上传您的简历，AI 将据此生成个性化问题"></el-step>
        <el-step title="选择面试配置" description="选择目标公司、岗位、面试轮次和难度等级"></el-step>
        <el-step title="开始面试" description="AI 面试官将依次提问，您可以随时暂停或继续"></el-step>
        <el-step title="查看报告" description="面试结束后获得详细评估报告，了解优势和不足"></el-step>
      </el-steps>
      <div class="start-action">
        <el-button type="primary" size="large" @click="$router.push('/interview/config')">
          🚀 立即开始面试
        </el-button>
      </div>
    </div>

    <!-- 功能详解 -->
    <div class="help-section">
      <h2 class="section-title">📚 功能详解</h2>
      <el-collapse v-model="activeCollapse" accordion>
        <el-collapse-item title="面试功能" name="interview">
          <template slot="title">
            <span class="collapse-title">🎤 面试功能</span>
          </template>
          <div class="collapse-content">
            <ul>
              <li><strong>发起面试</strong>：配置公司、岗位、轮次后，AI 生成个性化面试</li>
              <li><strong>视频面试准备</strong>：模拟真实视频面试环境，提前适应</li>
              <li><strong>面试历史</strong>：回顾过往面试表现，追踪成长轨迹</li>
              <li><strong>报告分享</strong>：生成面试报告链接，方便分享给他人</li>
            </ul>
          </div>
        </el-collapse-item>
        
        <el-collapse-item title="简历管理" name="resume">
          <template slot="title">
            <span class="collapse-title">📄 简历管理</span>
          </template>
          <div class="collapse-content">
            <ul>
              <li><strong>上传简历</strong>：支持 PDF、图片等多种格式</li>
              <li><strong>智能解析</strong>：自动提取关键信息，生成个性化问题</li>
              <li><strong>简历对比</strong>：分析简历与目标岗位的匹配度</li>
            </ul>
          </div>
        </el-collapse-item>
        
        <el-collapse-item title="题库系统" name="questions">
          <template slot="title">
            <span class="collapse-title">📝 题库系统</span>
          </template>
          <div class="collapse-content">
            <ul>
              <li><strong>分类浏览</strong>：按公司、岗位、难度筛选题目</li>
              <li><strong>单题练习</strong>：针对特定题目进行强化训练</li>
              <li><strong>贡献题目</strong>：分享您遇到的真实面试题，丰富题库</li>
              <li><strong>我的贡献</strong>：查看您贡献的题目及其采用情况</li>
            </ul>
          </div>
        </el-collapse-item>
        
        <el-collapse-item title="公司知识库" name="company">
          <template slot="title">
            <span class="collapse-title">🏢 公司知识库</span>
          </template>
          <div class="collapse-content">
            <ul>
              <li><strong>公司信息</strong>：了解目标公司的背景、文化、薪资范围</li>
              <li><strong>面试经验</strong>：查看其他用户的真实面试经验</li>
              <li><strong>常见问题</strong>：掌握高频面试问题，提前准备答案</li>
            </ul>
          </div>
        </el-collapse-item>
        
        <el-collapse-item title="知识社区" name="community">
          <template slot="title">
            <span class="collapse-title">🌐 知识社区</span>
          </template>
          <div class="collapse-content">
            <ul>
              <li><strong>经验分享</strong>：阅读和发布求职经验文章</li>
              <li><strong>互动交流</strong>：与其他求职者讨论面试技巧</li>
              <li><strong>信息获取</strong>：获取最新的招聘信息和行业动态</li>
            </ul>
          </div>
        </el-collapse-item>
        
        <el-collapse-item title="个人中心" name="profile">
          <template slot="title">
            <span class="collapse-title">👤 个人中心</span>
          </template>
          <div class="collapse-content">
            <ul>
              <li><strong>账户管理</strong>：修改个人信息、密码等</li>
              <li><strong>成长统计</strong>：查看面试次数、得分趋势等数据</li>
              <li><strong>设置偏好</strong>：自定义面试提醒、主题设置等</li>
            </ul>
          </div>
        </el-collapse-item>
      </el-collapse>
    </div>

    <!-- 常见问题 -->
    <div class="help-section">
      <h2 class="section-title">❓ 常见问题</h2>
      <div class="faq-list">
        <div class="faq-item" v-for="(faq, index) in faqs" :key="index">
          <div class="faq-question" @click="toggleFaq(index)">
            <span>{{ faq.q }}</span>
            <span class="faq-arrow" :class="{ 'faq-arrow-open': openFaq === index }">▼</span>
          </div>
          <div class="faq-answer" v-show="openFaq === index">
            {{ faq.a }}
          </div>
        </div>
      </div>
    </div>

    <!-- 底部提示 -->
    <div class="help-footer">
      <p>还有其他问题？<el-button type="text" @click="$router.push('/community')">前往社区</el-button>与其他用户交流，或联系客服。</p>
    </div>
  </div>
</template>

<script>
export default {
  name: 'Help',
  data() {
    return {
      currentStep: 0,
      activeCollapse: 'interview',
      openFaq: null,
      faqs: [
        {
          q: '面试需要多长时间？',
          a: '标准面试时长约15-30分钟，您可以在配置时选择面试长度，也可以随时暂停和继续。'
        },
        {
          q: 'AI 面试官如何评分？',
          a: '系统会从回答内容、逻辑表达、专业深度、时间控制等多个维度进行评估，最终给出综合得分和改进建议。'
        },
        {
          q: '我的简历会被泄露吗？',
          a: '您的简历仅用于生成个性化面试问题，不会被分享给任何第三方。我们承诺严格保护您的隐私。'
        },
        {
          q: '如何获得更准确的评估？',
          a: '建议上传完整的简历、选择匹配的岗位、认真回答每一个问题。多次练习后，系统会给出更准确的评估。'
        },
        {
          q: '面试报告可以分享吗？',
          a: '是的，您可以生成报告分享链接，但链接具有时效性，且仅能看到您选择公开的内容。'
        }
      ]
    }
  },
  methods: {
    toggleFaq(index) {
      this.openFaq = this.openFaq === index ? null : index
    }
  }
}
</script>

<style scoped>
.help-page {
  max-width: 1000px;
  margin: 0 auto;
  padding: 20px;
}

/* 顶部标题 */
.help-header {
  background: linear-gradient(135deg, #131921 0%, #232f3e 60%, #3a4a5c 100%);
  border-radius: 12px;
  padding: 40px;
  margin-bottom: 30px;
  text-align: center;
}
.header-title {
  font-size: 28px;
  color: #fff;
  margin-bottom: 10px;
}
.header-sub {
  font-size: 16px;
  color: #aab7c4;
  margin: 0;
}

/* 通用区块 */
.help-section {
  background: #fff;
  border-radius: 12px;
  padding: 24px;
  margin-bottom: 20px;
  border: 1px solid #eee;
}
.section-title {
  font-size: 18px;
  color: #111;
  margin: 0 0 20px 0;
  padding-bottom: 12px;
  border-bottom: 2px solid #ff9900;
  display: inline-block;
}

/* 功能卡片 */
.feature-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
}
.feature-card {
  display: flex;
  align-items: flex-start;
  gap: 16px;
  padding: 20px;
  border: 1px solid #eee;
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.2s;
}
.feature-card:hover {
  border-color: #ff9900;
  box-shadow: 0 4px 12px rgba(255, 153, 0, 0.15);
  transform: translateY(-2px);
}
.feature-icon {
  font-size: 36px;
  flex-shrink: 0;
}
.feature-info h3 {
  font-size: 16px;
  color: #111;
  margin: 0 0 8px 0;
}
.feature-info p {
  font-size: 13px;
  color: #666;
  margin: 0;
  line-height: 1.6;
}

/* 步骤条 */
.start-steps {
  margin: 30px 0;
}
.start-action {
  text-align: center;
  margin-top: 30px;
}

/* 折叠面板 */
.collapse-title {
  font-size: 15px;
  font-weight: 500;
}
.collapse-content ul {
  margin: 0;
  padding-left: 20px;
}
.collapse-content li {
  font-size: 14px;
  color: #555;
  line-height: 2;
}
.collapse-content strong {
  color: #333;
}

/* FAQ */
.faq-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.faq-item {
  border: 1px solid #eee;
  border-radius: 8px;
  overflow: hidden;
}
.faq-question {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 14px 16px;
  background: #fafafa;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  color: #333;
  transition: background 0.2s;
}
.faq-question:hover {
  background: #f5f5f5;
}
.faq-arrow {
  font-size: 12px;
  color: #999;
  transition: transform 0.3s;
}
.faq-arrow-open {
  transform: rotate(180deg);
}
.faq-answer {
  padding: 14px 16px;
  font-size: 14px;
  color: #666;
  line-height: 1.7;
  background: #fff;
}

/* 底部 */
.help-footer {
  text-align: center;
  padding: 20px;
  color: #888;
  font-size: 14px;
}
.help-footer p {
  margin: 0;
}

/* 响应式 */
@media (max-width: 768px) {
  .feature-grid {
    grid-template-columns: 1fr;
  }
  .help-header {
    padding: 30px 20px;
  }
  .header-title {
    font-size: 22px;
  }
}
</style>
