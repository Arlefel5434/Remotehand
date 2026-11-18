(() => {
  const status = document.querySelector("#status");
  const message = document.querySelector("#message");
  const form = document.querySelector("#text-form");
  const input = document.querySelector("#text-input");
  const controls = [...document.querySelectorAll("button")];
  const encoder = new TextEncoder();
  const mute = document.querySelector("#mute");
  const level = document.querySelector("#level");
  let muted = false;
  mute.onclick = () => send("mute", { muted: !muted });
  document.querySelectorAll("[data-volume]").forEach(button => {
    button.onclick = () => send("volume", { delta: Number(button.dataset.volume) });
  });

  let socket;
  let reconnectAttempt = 0;
  let reconnectTimer;
  const pad = document.querySelector("#pad");
  let gesture = null;
  let frame = 0;
  function resetGesture() {
    cancelAnimationFrame(frame);
    frame = 0;
    gesture = null;
  }
  function flushMotion() {
    frame = 0;
    if (!gesture || !gesture.moved) return;
    const dx = Math.round(gesture.x - gesture.sentX);
    const dy = Math.round(gesture.y - gesture.sentY);
    if (dx || dy) {
      const limit = n => Math.max(-200, Math.min(200, n));
      send("move", { dx: limit(dx), dy: limit(dy) });
      gesture.sentX += dx;
      gesture.sentY += dy;
    }
  }
  function track(event) {
    gesture.x = event.clientX;
    gesture.y = event.clientY;
    if (Math.hypot(gesture.x - gesture.startX, gesture.y - gesture.startY) > 4) gesture.moved = true;
  }
  pad.onpointerdown = event => {
    if (!event.isPrimary) { resetGesture(); return; }
    if (event.button !== 0 || socket?.readyState !== WebSocket.OPEN) return;
    resetGesture();
    pad.setPointerCapture(event.pointerId);
    gesture = { id: event.pointerId, startX: event.clientX, startY: event.clientY,
      sentX: event.clientX, sentY: event.clientY, moved: false, startTime: event.timeStamp };
  };
  pad.onpointermove = event => {
    if (gesture?.id !== event.pointerId) return;
    track(event);
    if (!frame && gesture.moved) frame = requestAnimationFrame(flushMotion);
  };
  pad.onpointerup = event => {
    if (gesture?.id !== event.pointerId) return;
    track(event);
    cancelAnimationFrame(frame);
    if (gesture.moved) flushMotion();
    else if (event.timeStamp - gesture.startTime <= 300) send("click", { button: "left" });
    resetGesture();
  };
  pad.onpointercancel = pad.onlostpointercapture = event => {
    if (gesture?.id === event.pointerId) resetGesture();
  };
  document.querySelectorAll("[data-button]").forEach(button => {
    button.onclick = () => send("click", { button: button.dataset.button });
  });

  function setConnected(connected) {
    resetGesture();
    status.textContent = connected ? "Подключено" : "Нет связи";
    status.classList.toggle("connected", connected);
    controls.forEach((control) => { control.disabled = !connected; });
    mute.disabled = true;
    level.textContent = "—";
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
      if (response.type === "volume_state") {
        muted = response.payload.muted;
        level.textContent = response.payload.level + "%";
        mute.textContent = muted ? "Включить звук" : "Выключить звук";
        mute.disabled = false;
      }
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
