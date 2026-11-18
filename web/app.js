(() => {
  const status = document.querySelector("#status");
  const message = document.querySelector("#message");
  const form = document.querySelector("#text-form");
  const input = document.querySelector("#text-input");
  const controls = [...document.querySelectorAll("button")];
  const encoder = new TextEncoder();

  let socket;
  let reconnectAttempt = 0;
  let reconnectTimer;

  function setConnected(connected) {
    status.textContent = connected ? "Подключено" : "Нет связи";
    status.classList.toggle("connected", connected);
    controls.forEach((control) => { control.disabled = !connected; });
  }

  function connect() {
    clearTimeout(reconnectTimer);
    if (socket && [WebSocket.OPEN, WebSocket.CONNECTING].includes(socket.readyState)) return;

    status.textContent = "Подключение…";
    const scheme = location.protocol === "https:" ? "wss:" : "ws:";
    socket = new WebSocket(`${scheme}//${location.host}/ws`);

    socket.addEventListener("open", () => {
      reconnectAttempt = 0;
      message.textContent = "";
      setConnected(true);
    });
    socket.addEventListener("close", () => {
      setConnected(false);
      const delay = Math.min(500 * (2 ** reconnectAttempt), 10000);
      reconnectAttempt += 1;
      reconnectTimer = setTimeout(connect, delay);
    });
    socket.addEventListener("error", () => socket.close());
    socket.addEventListener("message", (event) => {
      const response = JSON.parse(event.data);
      if (response.type === "error") message.textContent = response.payload.message;
    });
  }

  function send(type, payload) {
    if (!socket || socket.readyState !== WebSocket.OPEN) {
      message.textContent = "Нет связи с ноутбуком";
      return false;
    }
    socket.send(JSON.stringify({
      version: 1,
      type,
      id: commandID(),
      payload,
    }));
    return true;
  }

  function commandID() {
    if (crypto.randomUUID) return crypto.randomUUID();
    return `${Date.now().toString(36)}-${Math.random().toString(36).slice(2)}`;
  }

  form.addEventListener("submit", (event) => {
    event.preventDefault();
    const value = input.value;
    if (!value) return;
    if (encoder.encode(value).byteLength > 2048) {
      message.textContent = "Текст должен занимать не более 2048 байт";
      return;
    }
    if (send("text", { value })) {
      input.value = "";
      input.focus();
    }
  });

  document.querySelectorAll("[data-key]").forEach((button) => {
    button.addEventListener("click", () => send("key", { key: button.dataset.key }));
  });
  window.addEventListener("online", connect);
  setConnected(false);
  connect();
})();
