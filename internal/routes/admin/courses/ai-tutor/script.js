function sendPrompt(text) {
  document.getElementById('chat-input').value = text;
  sendUserMsg();
}

function sendUserMsg() {
  const input = document.getElementById('chat-input');
  const txt = input.value.trim();
  if (!txt) return;

  const msgContainer = document.getElementById('chat-messages');
  msgContainer.innerHTML += `<div class="msg msg-user">${txt}</div>`;
  input.value = '';
  msgContainer.scrollTop = msgContainer.scrollHeight;

  // AI response simulator
  setTimeout(() => {
    let reply = "Let's diagnose this systematically. Check your VDD_MAIN and VDD_BOOST rails first using your multimeter in diode mode with the red probe on ground.";
    const lower = txt.toLowerCase();
    if (lower.includes('0.05a') || lower.includes('stuck')) {
      reply = "A 0.05A (50mA) stuck draw on Apple boards typically signifies an uninitialized CPU/NAND clock sequence. Check the 24MHz master crystal oscillator output and ensure I2C0 bus lines are not pulled low by a cracked audio IC or sensor.";
    } else if (lower.includes('c1023') || lower.includes('diode')) {
      reply = "On a functional iPhone 13 board, capacitor C1023 on the PP_VDD_MAIN line should read approximately 0.380V - 0.395V in diode mode (Red probe on Ground). If you measure 0.000V, you have a direct short to ground.";
    } else if (lower.includes('reball') || lower.includes('temp')) {
      reply = "For PMIC BGA reballing using 183°C medium-melt paste: preheat the board to 180°C on the heating platform, set the hot air rework wand to 330°C - 340°C with 40% airflow, and circle gently from 2cm away until all spheres liquefy.";
    }

    msgContainer.innerHTML += `<div class="msg msg-ai">${reply}</div>`;
    msgContainer.scrollTop = msgContainer.scrollHeight;
  }, 600);
}

function resetChat() {
  document.getElementById('chat-messages').innerHTML = `
    <div class="msg msg-ai">
      Hello! I am your 24/7 Salen Mobile Guru Hardware Diagnostic Mentor. Ask me anything about diode mode test points, power sequences, BGA reballing curves, or short circuit isolation.
    </div>
  `;
}

function saveConfig() {
  alert('AI Mentor settings saved successfully!');
}
