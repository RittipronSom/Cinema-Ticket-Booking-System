<script setup>
import { ref, onMounted } from "vue";
import { auth, loginWithGoogle, logout } from "../firebase";
import { onAuthStateChanged } from "firebase/auth";

const seats = ref([]);
const user = ref(null);
const token = ref(null);
const lockedByMe = ref(
  JSON.parse(sessionStorage.getItem("lockedByMe") || "[]"),
);

onAuthStateChanged(auth, async (firebaseUser) => {
  if (firebaseUser) {
    user.value = firebaseUser;
    token.value = await firebaseUser.getIdToken();
  } else {
    user.value = null;
    token.value = null;
  }
});

const fetchSeats = async () => {
  const response = await fetch("http://localhost:8080/seats");
  seats.value = await response.json();
};

onMounted(() => {
  fetchSeats();
  const socket = new WebSocket("ws://localhost:8080/ws");
  socket.onmessage = () => fetchSeats();
});

const getSeatClass = (status) => {
  if (status === "AVAILABLE") return "available";
  if (status === "BOOKED") return "booked";
  if (status === "LOCKED") return "locked";
};

const lockSeat = async (seatNumber) => {
  if (!user.value) {
    alert("กรุณา Login ก่อน");
    return;
  }

  // ถ้าเป็นที่นั่งที่ตัวเองล็อคไว้ → ยกเลิก
  if (lockedByMe.value.includes(seatNumber)) {
    await fetch("http://localhost:8080/unlock-seat", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: "Bearer " + token.value,
      },
      body: JSON.stringify({ seat_number: seatNumber }),
    });
    lockedByMe.value = lockedByMe.value.filter((s) => s !== seatNumber);
    sessionStorage.setItem("lockedByMe", JSON.stringify(lockedByMe.value));
    fetchSeats();
    return;
  }

  // ถ้าเป็นที่นั่ง AVAILABLE → lock ปกติ
  const response = await fetch("http://localhost:8080/lock-seat", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Authorization: "Bearer " + token.value,
    },
    body: JSON.stringify({ seat_number: seatNumber }),
  });

  if (response.ok) {
    lockedByMe.value.push(seatNumber);
    sessionStorage.setItem("lockedByMe", JSON.stringify(lockedByMe.value));
  }

  fetchSeats();
};

const confirmAll = async () => {
  for (const seatNumber of lockedByMe.value) {
    await fetch("http://localhost:8080/confirm-booking", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: "Bearer " + token.value,
      },
      body: JSON.stringify({
        user_id: user.value.email,
        seat_number: seatNumber,
      }),
    });
  }
  lockedByMe.value = [];
  sessionStorage.setItem("lockedByMe", JSON.stringify([]));
  fetchSeats();
  alert("จองสำเร็จ!");
};
</script>

<template>
  <div class="page">
    <header class="header">
      <div class="brand">CINEMA</div>

      <div class="actions">
        <template v-if="user">
          <span class="user-name">{{ user.displayName }}</span>
          <button
            v-if="user.email === 'foj3010@gmail.com'"
            class="btn btn-ghost"
            @click="$router.push('/admin')"
          >
            Admin
          </button>
          <button class="btn btn-ghost" @click="logout">Logout</button>
        </template>
        <button v-else class="btn btn-solid" @click="loginWithGoogle">
          Sign in
        </button>
      </div>
    </header>

    <main class="main">
      <div class="intro">
        <span class="eyebrow">Now Showing</span>
        <h1 class="title">Select Your Seat</h1>
        <p class="subtitle">
          Tap an available seat to reserve it for 5 minutes.
        </p>
      </div>

      <div class="legend">
        <div class="legend-item">
          <span class="dot available"></span>
          <span>Available</span>
        </div>
        <div class="legend-item">
          <span class="dot locked"></span>
          <span>Locked</span>
        </div>
        <div class="legend-item">
          <span class="dot booked"></span>
          <span>Booked</span>
        </div>
      </div>

      <div class="screen">SCREEN</div>

      <div class="seat-grid">
        <div
          v-for="seat in seats"
          :key="seat.id"
          class="seat"
          :class="getSeatClass(seat.status)"
          @click="lockSeat(seat.seat_number)"
        >
          <div class="seat-number">{{ seat.seat_number }}</div>
            <div class="seat-status">{{ seat.status }}</div>
          
        </div>
      </div>
      <div class="confirm-bar">
        <span class="confirm-seats">
          ที่นั่งที่เลือก:
          <strong>{{
            lockedByMe.length > 0 ? lockedByMe.join(", ") : "ไม่มีที่นั่ง"
          }}</strong>
        </span>
        <button
          class="btn btn-confirm"
          :disabled="lockedByMe.length === 0"
          @click="confirmAll"
        >
          Confirm
          {{ lockedByMe.length > 0 ? lockedByMe.length + " ที่นั่ง" : "" }}
        </button>
      </div>
    </main>
  </div>
</template>

<style>
:root {
  --bg: #ffffff;
  --bg-soft: #f5f7fb;
  --fg: #0b1b3a;
  --fg-muted: #6b7a99;
  --fg-soft: #9aa6c0;
  --blue: #1a4ed8;
  --blue-deep: #0f3bb8;
  --blue-soft: #d6e3ff;
  --blue-pale: #f0f5ff;
  --border: #e3e8f2;
}

.page {
  height: 100vh;
  overflow: hidden;
  background: var(--bg);
  color: var(--fg);
  display: flex;
  flex-direction: column;
}

.header {
  position: relative;
  z-index: 10;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 28px 64px;
  border-bottom: 1px solid var(--border);
  background: var(--bg);
}

.brand {
  font-size: 22px;
  font-weight: 800;
  letter-spacing: 0.2em;
  color: var(--blue);
}

.actions {
  display: flex;
  align-items: center;
  gap: 16px;
}

.user-name {
  font-size: 16px;
  font-weight: 500;
  color: var(--fg-muted);
}

.btn {
  font-size: 15px;
  font-weight: 600;
  letter-spacing: 0.02em;
  padding: 10px 22px;
  border-radius: 999px;
  transition: all 0.2s ease;
  border: 1px solid transparent;
  display: inline-block;
}

.btn-solid {
  background: var(--blue);
  color: #fff;
}
.btn-solid:hover {
  background: var(--blue-deep);
  transform: translateY(-1px);
}

.btn-ghost {
  background: transparent;
  color: var(--blue);
  border: 1px solid var(--blue-soft);
}
.btn-ghost:hover {
  background: var(--blue-pale);
  border-color: var(--blue);
}

.btn-confirm {
  background: var(--blue);
  color: #fff;
  border-radius: 999px;
  padding: 8px 18px;
  font-size: 13px;
  font-weight: 600;
  letter-spacing: 0.05em;
}
.btn-confirm:hover {
  background: var(--blue-deep);
}
.btn-confirm:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}
body {
  overflow: hidden;
  height: 100vh;
}
.main {
  flex: 1;
  max-width: 900px;
  margin: 0 auto;
  padding: 50px 32px 20px; /* top right bottom */
  width: 100%;
}

.intro {
  text-align: center;
  margin-bottom: 56px;
}

.eyebrow {
  display: inline-block;
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.3em;
  text-transform: uppercase;
  color: var(--blue);
  background: var(--blue-pale);
  padding: 6px 16px;
  border-radius: 999px;
  margin-bottom: 20px;
}

.title {
  font-size: 56px;
  font-weight: 800;
  letter-spacing: -0.03em;
  margin-bottom: 12px;
  color: var(--fg);
  line-height: 1.1;
}

.subtitle {
  font-size: 18px;
  color: var(--fg-muted);
  max-width: 520px;
  margin: 0 auto;
}
.confirm-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  max-width: 540px;
  margin: 24px auto 0;
  padding: 16px 24px;
  background: var(--blue-pale);
  border: 1px solid var(--blue-soft);
  border-radius: 12px;
}

.confirm-seats {
  font-size: 15px;
  color: var(--fg);
}

.legend {
  display: flex;
  justify-content: center;
  gap: 36px;
  margin-bottom: 56px;
  font-size: 15px;
  color: var(--fg-muted);
  font-weight: 500;
}

.legend-item {
  display: flex;
  align-items: center;
  gap: 10px;
}

.dot {
  width: 16px;
  height: 16px;
  border-radius: 4px;
  border: 2px solid var(--blue);
  background: var(--bg);
  flex-shrink: 0;
}
.dot.locked {
  background: var(--blue);
}
.dot.booked {
  background: var(--blue-pale);
  border: 2px solid var(--blue-soft);
  opacity: 0.6;
}

.legend-item span:not(.dot) {
  line-height: 1;
}

.screen {
  text-align: center;
  font-size: 13px;
  font-weight: 700;
  letter-spacing: 0.5em;
  color: var(--blue);
  border-top: 2px solid var(--blue);
  margin: 0 auto 64px;
  padding-top: 18px;
  width: 80%;
  max-width: 540px;
}

.seat-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 20px;
  max-width: 540px;
  margin: 0 auto;
}

.seat {
  aspect-ratio: 1 / 1;
  border: 2px solid var(--blue);
  border-radius: 12px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  background: var(--bg);
  color: var(--blue);
  transition: all 0.2s ease;
  user-select: none;
}

.seat:hover.available {
  background: var(--blue);
  color: #fff;
  transform: translateY(-2px);
  box-shadow: 0 8px 20px rgba(26, 78, 216, 0.25);
}

.seat-number {
  font-size: 20px;
  font-weight: 700;
  letter-spacing: 0.02em;
}

.seat-status {
  font-size: 10px;
  font-weight: 600;
  letter-spacing: 0.15em;
  color: var(--fg-soft);
  margin-top: 4px;
  text-transform: uppercase;
}

.seat:hover.available .seat-status {
  color: #fff;
  opacity: 0.85;
}

.seat.locked {
  background: var(--blue);
  color: #fff;
  border-color: var(--blue);
}
.seat.locked .seat-status {
  color: #fff;
  opacity: 0.75;
}

.seat.booked {
  cursor: not-allowed;
  background: var(--blue-pale);
  border-color: var(--blue-soft);
  color: var(--fg-soft);
  opacity: 0.6;
}
.seat.booked .seat-status {
  color: var(--blue);
  opacity: 0.7;
}

@media (max-width: 640px) {
  .header {
    padding: 20px 24px;
  }
  .brand {
    font-size: 18px;
  }
  .main {
    padding: 56px 16px;
  }
  .title {
    font-size: 36px;
  }
  .subtitle {
    font-size: 16px;
  }
  .seat-grid {
    gap: 12px;
  }
  .seat-number {
    font-size: 16px;
  }
}
</style>
