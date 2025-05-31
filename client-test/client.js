class NotificationApp {
  constructor() {
    this.user = null;
    this.socket = null;
    this.notificationsList = document.getElementById("notifications");
    this.userLabel = document.getElementById("user-label");
    this.sendForm = document.getElementById("send-form");
    this.connectionStatus = document.getElementById("connection-status");
    this.showUnreadOnly = false;

    this.init();
  }

  async init() {
    await this.promptForUsername();
    this.setupEventListeners();
    this.addFilterButtons();
    await this.loadHistoricalNotifications();
    this.connectWebSocket();
  }

  async promptForUsername() {
    this.user = prompt("Insert your username:");
    if (!this.user) {
      alert("Username is required!");
      return this.promptForUsername();
    }
    this.userLabel.textContent = this.user;
  }

  setupEventListeners() {
    this.sendForm.addEventListener("submit", this.handleSendForm.bind(this));
  }

  async loadHistoricalNotifications() {
    try {
      this.updateConnectionStatus("connecting", "Loading notifications...");

      const url = new URL(`http://localhost:8080/notifications`);

      if (this.showUnreadOnly) {
        url.searchParams.append("unread", "true");
      }

      const response = await fetch(url, {
        method: "GET",
        headers: {
          "Content-Type": "application/json",
          "X-User-Username": this.user,
        },
      });

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }

      const notifications = await response.json();

      // Clear existing notifications
      this.notificationsList.innerHTML = "";

      if (!Array.isArray(notifications)) {
        return;
      }

      if (notifications.length !== 0) {
        notifications.forEach((notification) => {
          this.addNotificationToUI(notification);
        });
      }

      const filterText = this.showUnreadOnly ? "unread" : "all";
      console.log(`Loaded ${notifications.length} ${filterText} notifications`);
    } catch (error) {
      console.error("Error loading historical notifications:", error);
      this.showError("Failed to load notifications. Please refresh the page.");
    }
  }

  connectWebSocket() {
    try {
      this.updateConnectionStatus("connecting", "Connecting...");

      const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
      const wsUrl = `${protocol}//localhost:8080/ws?username=${encodeURIComponent(
        this.user
      )}`;

      this.socket = new WebSocket(wsUrl);

      this.socket.onopen = () => {
        console.log("WebSocket connected");
        this.updateConnectionStatus("connected", "Connected");
      };

      this.socket.onmessage = (event) => {
        try {
          const notification = JSON.parse(event.data);
          this.addNotificationToUI(notification);
          this.showNotificationAlert(notification);
        } catch (error) {
          console.error("Error parsing WebSocket message:", error);
        }
      };

      this.socket.onclose = (event) => {
        console.log("WebSocket disconnected:", event.code, event.reason);
        this.updateConnectionStatus("disconnected", "Disconnected");

        // Attempt to reconnect after 3 seconds
        setTimeout(() => {
          console.log("Attempting to reconnect...");
          this.connectWebSocket();
        }, 3000);
      };

      this.socket.onerror = (error) => {
        console.error("WebSocket error:", error);
        this.updateConnectionStatus("disconnected", "Connection Error");
      };
    } catch (error) {
      console.error("Failed to create WebSocket connection:", error);
      this.updateConnectionStatus("disconnected", "Failed to Connect");
    }
  }

  addFilterButtons() {
    const filterContainer = document.createElement("div");
    filterContainer.className = "filter-container";
    filterContainer.style.cssText = `
      margin-bottom: 15px;
      display: flex;
      gap: 10px;
    `;

    const allBtn = document.createElement("button");
    allBtn.textContent = "All";
    allBtn.className = "btn filter-btn active";
    allBtn.style.cssText = `
      padding: 8px 16px;
      font-size: 14px;
      background: #3498db;
      color: white;
      border: none;
      border-radius: 4px;
      cursor: pointer;
    `;
    const unreadBtn = document.createElement("button");
    unreadBtn.textContent = "Unread Only";
    unreadBtn.className = "btn filter-btn";
    unreadBtn.style.cssText = `
      padding: 8px 16px;
      font-size: 14px;
      background: #95a5a6;
      color: white;
      border: none;
      border-radius: 4px;
      cursor: pointer;
    `;

    allBtn.addEventListener("click", () => {
      this.showUnreadOnly = false;
      this.updateFilterButtons(allBtn, unreadBtn);
      this.loadHistoricalNotifications();
    });

    unreadBtn.addEventListener("click", () => {
      this.showUnreadOnly = true;
      this.updateFilterButtons(unreadBtn, allBtn);
      this.loadHistoricalNotifications();
    });
    filterContainer.appendChild(allBtn);
    filterContainer.appendChild(unreadBtn);

    const notificationsSection = document.querySelector(
      ".notifications-section"
    );
    notificationsSection.insertBefore(filterContainer, this.notificationsList);
  }

  updateFilterButtons(activeBtn, inactiveBtn) {
    activeBtn.style.background = "#3498db";
    inactiveBtn.style.background = "#95a5a6";
  }

  addNotificationToUI(notification) {
    const li = document.createElement("li");
    li.className = "notification-item";
    li.dataset.notificationId = notification.id;

    if (notification.read) {
      li.classList.add("read");
    }

    li.innerHTML = `
      <div class="notification-meta">
        From: <strong>${this.escapeHtml(notification.sender)}</strong> • 
        ${this.formatTimestamp(notification.createdAt)}
      </div>
      <div class="notification-message">${this.escapeHtml(
        notification.message
      )}</div>
    `;

    li.addEventListener("click", () => {
      this.markAsRead(notification.id, li);
    });

    // Add to the top of the list
    this.notificationsList.insertBefore(li, this.notificationsList.firstChild);
  }

  async markAsRead(notificationId, element) {
    if (element.classList.contains("read")) {
      return; // Already marked as read
    }

    try {
      const response = await fetch(
        `http://localhost:8080/notifications/${encodeURIComponent(
          notificationId
        )}`,
        {
          method: "PATCH",
          headers: {
            "Content-Type": "application/json",
            "X-User-Username": this.user,
          },
        }
      );

      if (response.ok) {
        element.classList.add("read");
        console.log(`Notification ${notificationId} marked as read`);
      } else {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }
    } catch (error) {
      console.error("Error marking notification as read:", error);
      this.showError("Failed to mark notification as read");
    }
  }

  async handleSendForm(event) {
    event.preventDefault();

    const receiverInput = document.getElementById("receiver");
    const messageInput = document.getElementById("message");

    const receiver = receiverInput.value.trim();
    const message = messageInput.value.trim();

    if (!receiver || !message) {
      this.showError("Please fill in both receiver and message");
      return;
    }

    try {
      const response = await fetch("http://localhost:8080/notifications", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "X-User-Username": this.user,
        },
        body: JSON.stringify({
          receiver: receiver,
          message: message,
        }),
      });

      if (response.ok) {
        messageInput.value = "";
        receiverInput.value = "";
        console.log("Notification sent successfully");
      } else {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }
    } catch (error) {
      console.error("Error sending notification:", error);
      this.showError("Failed to send notification");
    }
  }

  updateConnectionStatus(status, message) {
    this.connectionStatus.className = `connection-status ${status}`;
    this.connectionStatus.textContent = message;

    // Auto-hide connection status after 3 seconds if connected
    if (status === "connected") {
      setTimeout(() => {
        this.connectionStatus.style.opacity = "0.7";
      }, 3000);
    } else {
      this.connectionStatus.style.opacity = "1";
    }
  }

  showNotificationAlert(notification) {
    // Only show alert if the page is not visible
    if (document.hidden) {
      if (Notification.permission === "granted") {
        new Notification(`New message from ${notification.sender}`, {
          body: notification.message,
          icon: "/static/icon.png", // Add an icon if you have one
        });
      }
    }
  }

  showError(message) {
    // Simple error display - you could make this more sophisticated
    const errorDiv = document.createElement("div");
    errorDiv.className = "error-message";
    errorDiv.style.cssText = `
      position: fixed;
      top: 20px;
      left: 50%;
      transform: translateX(-50%);
      background: #e74c3c;
      color: white;
      padding: 12px 20px;
      border-radius: 6px;
      z-index: 1000;
      box-shadow: 0 4px 12px rgba(0,0,0,0.2);
    `;
    errorDiv.textContent = message;

    document.body.appendChild(errorDiv);

    setTimeout(() => {
      errorDiv.remove();
    }, 5000);
  }

  escapeHtml(text) {
    const div = document.createElement("div");
    div.textContent = text;
    return div.innerHTML;
  }

  formatTimestamp(timestamp) {
    if (!timestamp) return "Just now";

    const date = new Date(timestamp * 1000);
    const now = new Date();
    const diffMs = now - date;
    const diffMins = Math.floor(diffMs / 60000);
    const diffHours = Math.floor(diffMs / 3600000);
    const diffDays = Math.floor(diffMs / 86400000);

    if (diffMins < 1) return "Just now";
    if (diffMins < 60) return `${diffMins}m ago`;
    if (diffHours < 24) return `${diffHours}h ago`;
    if (diffDays < 7) return `${diffDays}d ago`;

    return date.toLocaleDateString();
  }
}

// Request notification permission
if ("Notification" in window && Notification.permission === "default") {
  Notification.requestPermission();
}

// Initialize the app when DOM is loaded
document.addEventListener("DOMContentLoaded", () => {
  new NotificationApp();
});
