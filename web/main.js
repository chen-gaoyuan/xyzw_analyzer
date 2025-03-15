layui.use(["layer", "element"], function () {
  var layer = layui.layer;
  var element = layui.element;
  var websocket;
  var currentJson = "";

  // Connect to WebSocket
  function connectWebSocket() {
    var wsUrl = "ws://127.0.0.1:12582/ws";

    if (websocket && websocket.readyState !== WebSocket.CLOSED) {
      websocket.close();
    }

    websocket = new WebSocket(wsUrl);

    websocket.onopen = function () {
      $("#status").text("已连接").removeClass("disconnected error").addClass("connected");
      layer.msg("WebSocket 连接成功", { icon: 1 });
    };

    websocket.onmessage = function (event) {
      var message = JSON.parse(event.data);
      displayMessage(message);
    };

    websocket.onclose = function () {
      $("#status").text("已断开").removeClass("connected error").addClass("disconnected");
      layer.msg("WebSocket 连接已关闭", { icon: 5 });
    };

    websocket.onerror = function () {
      $("#status").text("连接错误").removeClass("connected disconnected").addClass("error");
      layer.msg("WebSocket 连接错误", { icon: 5 });
    };
  }

  // 初始连接
  connectWebSocket();

  // 重新连接按钮
  $("#reconnectBtn").click(function() {
    connectWebSocket();
  });

  // 清空消息按钮
  $("#clearBtn").click(function() {
    $("#messages").empty();
    $("#jsonContent").text("");
    currentJson = "";
    layer.msg("消息已清空", { icon: 1 });
  });

  // 复制JSON按钮
  $("#copyJsonBtn").click(function() {
    copyToClipboard(currentJson);
  });

  // 复制到剪贴板函数
  function copyToClipboard(text) {
    if (!text) {
      layer.msg("没有内容可复制", { icon: 2 });
      return;
    }

    navigator.clipboard.writeText(text).then(function() {
      layer.msg("已复制到剪贴板", { icon: 1 });
    }, function() {
      layer.msg("复制失败，请手动复制", { icon: 2 });
    });
  }

  // Display JSON message
  function displayMessage(message) {
    try {
      var msg = JSON.parse(message.msg);
      var messageHtml = '<div class="message-item">';
      if (message.call == "client") {
        messageHtml +=
            '<div class="message-time red">' +
            getCurrentTime() +
            "   Send: " +
            msg.cmd +
            "</div>";
      } else {
        messageHtml +=
            '<div class="message-time blue">' +
            getCurrentTime() +
            "  Receive: " +
            msg.cmd +
            "</div>";
      }

      messageHtml += '<div class="message-content">';
      messageHtml += "<pre>" + JSON.stringify(msg, null, 2) + "</pre>";
      messageHtml += "</div>";

      if (message.call == "client") {
        messageHtml +=
            '<button class="layui-btn layui-btn-primary layui-btn-sm debug-btn">调试</button>';
      } else {
        messageHtml +=
            '<button class="layui-btn layui-btn-primary layui-btn-sm debug-btn">查看</button>';
      }
      messageHtml += "</div>";

      $("#messages").prepend(messageHtml); // New message at the top

      // Bind click event to new message item
      var newMessageItem = $("#messages .message-item:first-child");
      newMessageItem.click(function () {
        $(this).toggleClass("expanded");
      });

      // Bind click event to debug button
      newMessageItem.find(".debug-btn").click(function (event) {
        event.stopPropagation(); // Prevent message item from collapsing
        showJson(message); // Call function to show JSON
      });

      // Scroll to top
      var messageContainer = $("#messages");
      messageContainer.scrollTop(0);
    } catch (e) {
      console.error("Error parsing message:", e);
      layer.msg("消息解析错误: " + e.message, { icon: 2 });
    }
  }

  // Function to show JSON
  function showJson(message) {
    try {
      var formattedJson = JSON.stringify(JSON.parse(message.msg), null, 2);
      $("#jsonContent").text(formattedJson);
      currentJson = formattedJson;
    } catch (e) {
      console.error("Error showing JSON:", e);
      layer.msg("JSON解析错误: " + e.message, { icon: 2 });
    }
  }

  // Function to get current time in hh:mm:ss format
  function getCurrentTime() {
    var now = new Date();
    var hours = String(now.getHours()).padStart(2, '0');
    var minutes = String(now.getMinutes()).padStart(2, '0');
    var seconds = String(now.getSeconds()).padStart(2, '0');
    return hours + ":" + minutes + ":" + seconds;
  }
});