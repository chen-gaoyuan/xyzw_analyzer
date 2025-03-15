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
            filterDirection: 'all', // 'all', 'client', 'server'
            filterCmd: '',
            // 排除相关
            excludedCommands: [], // 存储被排除的命令列表
            // 备注相关
            commandNotes: {}, // 存储命令备注，格式: {cmd: 'note'}
            keyNotes: {},     // 存储键备注，格式: {cmd: {key: 'note'}}
            currentEditingNote: null, // 当前正在编辑的备注对象
            noteDialogVisible: false, // 备注对话框可见性
            noteContent: '',          // 备注内容
            // Tab相关
            activeTab: 'detail',
            debugContent: '',
            jsonTitle: ''
        };
    },
// 添加watch监听noteDialogVisible的变化
    watch: {
        noteDialogVisible(newVal) {
            // 当对话框关闭时，如果是键备注，重新渲染JSON视图
            if (!newVal && this.currentEditingNote && this.currentEditingNote.type === 'key') {
                this.$nextTick(() => {
                    this.addJsonKeyClickHandlers();
                });
            }
        }
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
        connectionButtonText() {
            if (this.isConnected) return '断开链接';
            return '重新链接';
        },
        // 修改filteredMessages计算属性，添加排除命令的逻辑
        filteredMessages() {
            return this.messages.filter(message => {
                // 排除指定命令
                if (this.excludedCommands.includes(message.parsedMsg.cmd)) {
                    return false;
                }

                // 方向筛选
                if (this.filterDirection !== 'all' && message.call !== this.filterDirection) {
                    return false;
                }

                // 命令筛选
                if (this.filterCmd && (!message.parsedMsg.cmd ||
                    !message.parsedMsg.cmd.toLowerCase().includes(this.filterCmd.toLowerCase()))) {
                    return false;
                }


                return true;
            });
        },


    },

    mounted() {
        // 组件挂载后自动连接WebSocket
        this.connectWebSocket();
        // 加载备注
        this.loadNotes();
        // 加载排除命令列表
        this.loadExcludedCommands();
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
            if (this.isConnected) {
                this.closeWebSocket();
            } else {
                this.connectWebSocket();
            }
        },

        // 处理接收到的消息
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


        // 查看消息详情时，设置调试内容
        viewMessageDetail(message) {
            try {
                this.currentMessage = message; // 保存当前查看的消息，用于键备注
                console.log(this.currentMessage)
                this.currentJson = this.formatJson(message.parsedMsg);
                this.jsonTitle = `${message.parsedMsg.cmd}   (${this.commandNotes[message.parsedMsg.cmd]})`;

                // 设置调试内容为body部分
                if (message.parsedMsg.body) {
                    this.debugContent = this.formatJson(message.parsedMsg.body);
                } else {
                    this.debugContent = '{}';
                }

                // 在DOM更新后添加JSON键的点击事件
                this.$nextTick(() => {
                    this.addJsonKeyClickHandlers();
                });
            } catch (e) {
                console.error('JSON格式化错误:', e);
                this.$message.error('JSON格式化错误: ' + e.message);
            }
        },
        // 格式化调试内容
        formatDebugContent() {
            try {
                const jsonObj = JSON.parse(this.debugContent);
                this.debugContent = this.formatJson(jsonObj);
                this.$message.success('格式化成功');
            } catch (e) {
                this.$message.error('JSON格式错误: ' + e.message);
            }
        },
        // 发送调试消息
        sendDebugMessage() {
            if (!this.isConnected) {
                this.$message.error('WebSocket未连接');
                return;
            }

            try {
                const jsonObj = {};
                jsonObj.cmd = this.currentMessage.parsedMsg.cmd;
                jsonObj.data=JSON.parse(this.debugContent)
                // 发送到后端队列
                fetch('/api/debug/send', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify(jsonObj)
                })
                    .then(response => {
                        if (!response.ok) {
                            throw new Error('发送失败');
                        }
                        return response.json();
                    })
                    .then(data => {
                        this.$message.success('消息已加入队列，将在2秒内发送');
                    })
                    .catch(error => {
                        this.$message.error('发送失败: ' + error.message);
                    });
            } catch (e) {
                this.$message.error('JSON格式错误: ' + e.message);
            }
        },

// 添加新方法，用于处理JSON键的点击事件
        addJsonKeyClickHandlers() {
            const jsonContent = document.querySelector('.json-content pre');
            if (!jsonContent) return;

            // 解析JSON并创建带有点击事件的DOM结构
            const jsonObj = this.currentMessage?.parsedMsg;
            if (!jsonObj) return;

            // 清空原有内容
            jsonContent.innerHTML = '';

            // 创建格式化的JSON视图
            const formattedView = this.createFormattedJsonView(jsonObj);
            jsonContent.appendChild(formattedView);
        },

// 创建格式化的JSON视图
        createFormattedJsonView(jsonObj, path = '', level = 0) {
            const container = document.createElement('div');
            container.className = 'json-object-container';

            const indent = '  '.repeat(level);
            const cmd = this.currentMessage?.parsedMsg?.cmd;

            // 添加开括号
            const openBrace = document.createElement('div');
            openBrace.textContent = '{';
            container.appendChild(openBrace);

            // 处理所有键值对
            const keys = Object.keys(jsonObj);
            keys.forEach((key, index) => {
                const keyPath = path ? `${path}.${key}` : key;
                const value = jsonObj[key];
                const hasNote = this.hasKeyNote(cmd, keyPath);
                const note = this.getKeyNote(cmd, keyPath);

                const row = document.createElement('div');
                row.className = 'json-row';
                row.style.marginLeft = '20px';

                // 键名
                const keySpan = document.createElement('span');
                keySpan.className = 'json-key';
                keySpan.textContent = `"${key}": `;
                row.appendChild(keySpan);

                // 值
                if (typeof value === 'object' && value !== null) {
                    // 递归处理嵌套对象
                    const nestedContainer = this.createFormattedJsonView(value, keyPath, level + 1);
                    row.appendChild(nestedContainer);
                } else {
                    const valueSpan = document.createElement('span');
                    valueSpan.className = typeof value === 'string' ? 'json-string' : 'json-value';
                    valueSpan.textContent = typeof value === 'string' ? `"${value}"` : `${value}`;
                    row.appendChild(valueSpan);
                }

                // 添加逗号
                if (index < keys.length - 1) {
                    const comma = document.createElement('span');
                    comma.textContent = ',';
                    row.appendChild(comma);
                }

                // 添加备注或添加备注按钮
                if (hasNote) {
                    const noteSpan = document.createElement('span');
                    noteSpan.className = 'json-note';
                    noteSpan.textContent = `/* ${note} */`;
                    noteSpan.title = "点击编辑备注";

                    // 添加点击事件
                    noteSpan.addEventListener('click', (e) => {
                        e.stopPropagation();
                        this.openKeyNoteDialog(this.currentMessage, keyPath);
                    });

                    row.appendChild(noteSpan);
                } else {
                    const addNoteSpan = document.createElement('span');
                    addNoteSpan.className = 'json-add-note';
                    addNoteSpan.textContent = '+';
                    addNoteSpan.title = "添加备注";

                    // 添加点击事件
                    addNoteSpan.addEventListener('click', (e) => {
                        e.stopPropagation();
                        this.openKeyNoteDialog(this.currentMessage, keyPath);
                    });

                    row.appendChild(addNoteSpan);
                }

                container.appendChild(row);
            });

            // 添加闭括号
            const closeBrace = document.createElement('div');
            closeBrace.textContent = '}';
            if (level > 0) {
                closeBrace.style.marginLeft = '  '.repeat(level - 1);
            }
            container.appendChild(closeBrace);

            return container;
        },
        // 格式化JSON并添加备注功能
        renderJsonWithNotes(json, path = '') {
            if (!json || typeof json !== 'object') return '';

            const cmd = this.currentMessage?.parsedMsg?.cmd;
            if (!cmd) return this.formatJson(json);

            let result = '{\n';
            const keys = Object.keys(json);

            keys.forEach((key, index) => {
                const value = json[key];
                const keyPath = path ? `${path}.${key}` : key;
                const hasNote = this.hasKeyNote(cmd, keyPath);
                const note = this.getKeyNote(cmd, keyPath);

                result += `  "${key}": `;

                if (typeof value === 'object' && value !== null) {
                    result += this.renderJsonWithNotes(value, keyPath);
                } else if (typeof value === 'string') {
                    result += `"${value}"`;
                } else {
                    result += value;
                }

                if (index < keys.length - 1) {
                    result += ',';
                }

                if (hasNote) {
                    result += ` /* ${note} */`;
                }

                result += '\n';
            });

            result += '}';
            return result;
        },
        // 添加新方法，用于渲染带HTML的JSON
        renderJsonWithNotesHtml(json) {
            if (!json) return '';

            const cmd = this.currentMessage?.parsedMsg?.cmd;
            if (!cmd) return this.formatJson(json);

            // 使用递归函数处理JSON
            const processObject = (obj, path = '') => {
                if (!obj || typeof obj !== 'object') return '';

                let result = '{\n';
                const keys = Object.keys(obj);

                keys.forEach((key, index) => {
                    const value = obj[key];
                    const keyPath = path ? `${path}.${key}` : key;
                    const hasNote = this.hasKeyNote(cmd, keyPath);
                    const note = this.getKeyNote(cmd, keyPath);

                    // 添加键名
                    result += `  <span class="json-key">"${key}"</span>: `;

                    // 处理值
                    if (typeof value === 'object' && value !== null) {
                        result += processObject(value, keyPath);
                    } else if (typeof value === 'string') {
                        result += `<span class="json-string">"${value}"</span>`;
                    } else {
                        result += `<span class="json-value">${value}</span>`;
                    }

                    if (index < keys.length - 1) {
                        result += ',';
                    }

                    // 添加备注
                    if (hasNote) {
                        result += ` <span class="json-note" @click="openKeyNoteDialog(currentMessage, '${keyPath}')">/* ${note} */</span>`;
                    } else {
                        result += ` <span class="json-add-note" @click="openKeyNoteDialog(currentMessage, '${keyPath}')">+</span>`;
                    }

                    result += '\n';
                });

                result += '}';
                return result;
            };

            return processObject(json);
        },
        // 重置筛选条件
        resetFilters() {

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
// 排除指定命令的消息
        excludeCommand(cmd) {
            if (!this.excludedCommands.includes(cmd)) {
                this.excludedCommands.push(cmd);
                // 保存到本地存储
                localStorage.setItem('excludedCommands', JSON.stringify(this.excludedCommands));
                this.$message.success(`已排除命令: ${cmd}`);
            }
        },

        // 取消排除命令
        includeCommand(cmd) {
            const index = this.excludedCommands.indexOf(cmd);
            if (index > -1) {
                this.excludedCommands.splice(index, 1);
                // 保存到本地存储
                localStorage.setItem('excludedCommands', JSON.stringify(this.excludedCommands));
                this.$message.success(`已取消排除命令: ${cmd}`);
            }
        },

        // 加载排除命令列表
        loadExcludedCommands() {
            const excludedCommands = localStorage.getItem('excludedCommands');
            if (excludedCommands) {
                this.excludedCommands = JSON.parse(excludedCommands);
            }
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
        },
        // 保存备注到本地存储
        // 修改saveNotes方法，将备注数据保存到后端
        saveNotes() {
            // 同时保存到本地存储作为备份
            localStorage.setItem('commandNotes', JSON.stringify(this.commandNotes));
            localStorage.setItem('keyNotes', JSON.stringify(this.keyNotes));

            // 发送到后端保存
            fetch('/api/notes/save', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    commandNotes: this.commandNotes,
                    keyNotes: this.keyNotes
                })
            })
                .then(response => {
                    if (!response.ok) {
                        throw new Error('保存备注失败');
                    }
                    return response.json();
                })
                .then(data => {
                    console.log('备注数据保存成功');
                })
                .catch(error => {
                    console.error('保存备注错误:', error);
                    this.$message.error('保存备注失败: ' + error.message);
                });
        },

        // 修改loadNotes方法，从后端加载备注数据
        loadNotes() {
            fetch('/api/notes/load')
                .then(response => {
                    if (!response.ok) {
                        throw new Error('加载备注失败');
                    }
                    return response.json();
                })
                .then(data => {
                    this.commandNotes = data.commandNotes || {};
                    this.keyNotes = data.keyNotes || {};
                    console.log('备注数据加载成功');
                })
                .catch(error => {
                    console.error('加载备注错误:', error);
                    this.$message.error('加载备注失败: ' + error.message);

                    // 加载失败时尝试从本地存储加载
                    const commandNotes = localStorage.getItem('commandNotes');
                    const keyNotes = localStorage.getItem('keyNotes');

                    if (commandNotes) {
                        this.commandNotes = JSON.parse(commandNotes);
                    }

                    if (keyNotes) {
                        this.keyNotes = JSON.parse(keyNotes);
                    }
                });
        },
        // 打开命令备注对话框
        openCommandNoteDialog(message) {
            const cmd = message.parsedMsg.cmd;
            this.currentEditingNote = {type: 'command', cmd};
            this.noteContent = this.commandNotes[cmd] || '';
            this.noteDialogVisible = true;
        },

        // 打开键备注对话框
        // 修改openKeyNoteDialog方法
        openKeyNoteDialog(message, key) {
            if (!message) {
                message = this.currentMessage;
            }

            if (!message) {
                this.$message.error('无法获取消息信息');
                return;
            }

            const cmd = message.parsedMsg.cmd;
            this.currentEditingNote = {type: 'key', cmd, key};

            // 初始化嵌套对象
            if (!this.keyNotes[cmd]) {
                this.keyNotes[cmd] = {};
            }

            this.noteContent = this.keyNotes[cmd][key] || '';
            this.noteDialogVisible = true;
        },

        // 保存备注
        // 修改saveNote方法
        saveNote() {
            if (!this.currentEditingNote) return;

            if (this.currentEditingNote.type === 'command') {
                const {cmd} = this.currentEditingNote;
                if (this.noteContent.trim()) {
                    this.commandNotes[cmd] = this.noteContent;
                } else {
                    delete this.commandNotes[cmd];
                }
            } else if (this.currentEditingNote.type === 'key') {
                const {cmd, key} = this.currentEditingNote;
                if (!this.keyNotes[cmd]) {
                    this.keyNotes[cmd] = {};
                }

                if (this.noteContent.trim()) {
                    this.keyNotes[cmd][key] = this.noteContent;
                } else {
                    delete this.keyNotes[cmd][key];
                    // 如果没有键备注了，清理空对象
                    if (Object.keys(this.keyNotes[cmd]).length === 0) {
                        delete this.keyNotes[cmd];
                    }
                }
            }

            // 保存到本地存储
            this.saveNotes();
            this.noteDialogVisible = false;
            this.$message.success('备注已保存');

            // 如果当前正在查看JSON详情，重新渲染JSON视图
            if (this.currentMessage) {
                this.$nextTick(() => {
                    this.addJsonKeyClickHandlers();
                });
            }
        },

        // 获取命令备注
        getCommandNote(cmd) {
            return this.commandNotes[cmd] || '';
        },

        // 获取键备注
        getKeyNote(cmd, key) {
            if (!this.keyNotes[cmd]) return '';
            return this.keyNotes[cmd][key] || '';
        },

        // 检查命令是否有备注
        hasCommandNote(cmd) {
            return !!this.commandNotes[cmd];
        },

        // 检查键是否有备注
        hasKeyNote(cmd, key) {
            if (!this.keyNotes[cmd]) return false;
            return !!this.keyNotes[cmd][key];
        },
        // 处理排除命令下拉菜单的选择
        handleExcludedCommand(command) {
            if (command === 'clear-all') {
                // 清除所有排除
                this.clearAllExcludes();
            } else {
                // 取消排除特定命令
                this.includeCommand(command);
            }
        },

        // 清除所有排除
        clearAllExcludes() {
            if (this.excludedCommands.length === 0) return;

            this.excludedCommands = [];
            localStorage.setItem('excludedCommands', JSON.stringify(this.excludedCommands));
            this.$message.success('已取消所有排除');
        },
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