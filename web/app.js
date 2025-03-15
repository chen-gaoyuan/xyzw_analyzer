// 创建Vue应用
const App = {
  data() {
    return {
      // WebSocket相关
      websocket: null,
      isConnected: false,
      connectionError: false,
      connecting: false,

      // 消息数据
      messages: [],
      currentJson: '',

      // 最大消息数量限制，防止内存占用过多
      maxMessages: 100,

      // 筛选相关
      filterKeyword: '',
      filterDirection: 'all', // 'all', 'client', 'server'
      filterCmd: ''
    };
  },
  
  computed: {
    // 连接状态类型
    connectionStatusType() {
      if (this.isConnected) return 'success';
      if (this.connectionError) return 'danger';
      return 'info';
    },

    // 连接状态文本
    connectionStatusText() {
      if (this.isConnected) return '已连接';
      if (this.connectionError) return '连接错误';
      return '已断开';
    },

    // 筛选后的消息列表
    filteredMessages() {
      return this.messages.filter(message => {
        // 方向筛选
        if (this.filterDirection !== 'all' && message.call !== this.filterDirection) {
          return false;
        }

        // 命令筛选
        if (this.filterCmd && (!message.parsedMsg.cmd ||
            !message.parsedMsg.cmd.toLowerCase().includes(this.filterCmd.toLowerCase()))) {
          return false;
        }

        // 关键词筛选
        if (this.filterKeyword) {
          const jsonStr = JSON.stringify(message.parsedMsg).toLowerCase();
          return jsonStr.includes(this.filterKeyword.toLowerCase());
        }

        return true;
      });
    }
  },
  
  mounted() {
    // 组件挂载后自动连接WebSocket
    this.connectWebSocket();
  },
  
  beforeUnmount() {
    // 组件卸载前关闭WebSocket连接
    this.closeWebSocket();
  },
  
  methods: {
    // 连接WebSocket
    connectWebSocket() {
      const wsUrl = "ws://127.0.0.1:12582/ws";
      
      // 如果已有连接，先关闭
      if (this.websocket && this.websocket.readyState !== WebSocket.CLOSED) {
        this.websocket.close();
      }
      
      this.connecting = true;
      this.websocket = new WebSocket(wsUrl);
      
      // 连接成功
      this.websocket.onopen = () => {
        this.isConnected = true;
        this.connectionError = false;
        this.connecting = false;
        this.$message.success('WebSocket 连接成功');
      };
      
      // 接收消息
      this.websocket.onmessage = (event) => {
        try {
          const message = JSON.parse(event.data);
          this.processMessage(message);
        } catch (e) {
          console.error('消息解析错误:', e);
          this.$message.error('消息解析错误: ' + e.message);
        }
      };
      
      // 连接关闭
      this.websocket.onclose = () => {
        this.isConnected = false;
        this.connectionError = false;
        this.connecting = false;
        this.$message.warning('WebSocket 连接已关闭');
      };
      
      // 连接错误
      this.websocket.onerror = () => {
        this.isConnected = false;
        this.connectionError = true;
        this.connecting = false;
        this.$message.error('WebSocket 连接错误');
      };
    },
    
    // 关闭WebSocket连接
    closeWebSocket() {
      if (this.websocket && this.websocket.readyState !== WebSocket.CLOSED) {
        this.websocket.close();
      }
    },
    
    // 重新连接
    reconnect() {
      this.connectWebSocket();
    },
    
    // 处理接收到的消息
    processMessage(message) {
      try {
        // 解析消息内容
        const parsedMsg = JSON.parse(message.msg);
        
        // 创建新消息对象
        const newMessage = {
          call: message.call,
          msg: message.msg,
          parsedMsg: parsedMsg,
          timestamp: Date.now(),
          expanded: false
        };
        
        // 添加到消息列表开头
        this.messages.unshift(newMessage);
        
        // 限制消息数量
        if (this.messages.length > this.maxMessages) {
          this.messages = this.messages.slice(0, this.maxMessages);
        }
      } catch (e) {
        console.error('消息处理错误:', e);
        this.$message.error('消息处理错误: ' + e.message);
      }
    },
    
    // 切换消息展开/折叠状态
    toggleExpand(message) {
      message.expanded = !message.expanded;
    },
    
    // 查看消息详情
    viewMessageDetail(message) {
      try {
        this.currentJson = this.formatJson(message.parsedMsg);
      } catch (e) {
        console.error('JSON格式化错误:', e);
        this.$message.error('JSON格式化错误: ' + e.message);
      }
    },
    // 重置筛选条件
    resetFilters() {
      this.filterKeyword = '';
      this.filterDirection = 'all';
      this.filterCmd = '';
      this.$message.success('筛选条件已重置');
    },
    // 清空所有消息
    clearMessages() {
      this.messages = [];
      this.currentJson = '';
      this.$message.success('消息已清空');
    },
    
    // 复制JSON到剪贴板
    copyJson() {
      if (!this.currentJson) {
        this.$message.warning('没有内容可复制');
        return;
      }
      
      navigator.clipboard.writeText(this.currentJson)
        .then(() => {
          this.$message.success('已复制到剪贴板');
        })
        .catch(() => {
          this.$message.error('复制失败，请手动复制');
        });
    },
    
    // 格式化JSON
    formatJson(json) {
      return JSON.stringify(json, null, 2);
    },
    
    // 获取格式化的时间
    getCurrentTime(timestamp) {
      const date = timestamp ? new Date(timestamp) : new Date();
      const hours = String(date.getHours()).padStart(2, '0');
      const minutes = String(date.getMinutes()).padStart(2, '0');
      const seconds = String(date.getSeconds()).padStart(2, '0');
      return `${hours}:${minutes}:${seconds}`;
    }
  }
};

// 创建Vue应用并挂载
const app = Vue.createApp(App);

// 注册Element Plus图标
for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(key, component);
}

// 使用Element Plus
app.use(ElementPlus);

// 挂载应用
app.mount('#app');