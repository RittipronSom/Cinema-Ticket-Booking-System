<script setup>
import { ref, onMounted } from "vue";
import { auth, loginWithGoogle, logout } from "../firebase";
import { onAuthStateChanged } from "firebase/auth";

const seats = ref([]);
const user = ref(null);
const token = ref(null);
const lockedByMe = ref([]);

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
  }

  fetchSeats();
};

const confirmBooking = async (seatNumber) => {
  const response = await fetch("http://localhost:8080/confirm-booking", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Authorization: "Bearer " + token.value,
    },
    body: JSON.stringify({
      user_id: user.value.email, // เปลี่ยนจาก uid เป็น email
      seat_number: seatNumber,
    }),
  });

  const data = await response.json();
  alert(data.message || data.error);
  lockedByMe.value = lockedByMe.value.filter((s) => s !== seatNumber);
  fetchSeats();
};
</script>

<template>
  <div class="container">
    <div class="header">
      <h1>Cinema Seat Map</h1>

      <div v-if="user">
        <span>{{ user.displayName }}</span>
        <button
          v-if="user.email === 'foj3010@gmail.com'"
          @click="$router.push('/admin')"
        >
          Admin
        </button>
        <button @click="logout">Logout</button>
      </div>

      <button v-else @click="loginWithGoogle">Login with Google</button>
    </div>

    <div class="seat-grid">
      <div
        v-for="seat in seats"
        :key="seat.id"
        class="seat"
        :class="getSeatClass(seat.status)"
        @click="lockSeat(seat.seat_number)"
      >
        <div>{{ seat.seat_number }}</div>
        <div>{{ seat.status }}</div>

        <button
          v-if="
            seat.status === 'LOCKED' && lockedByMe.includes(seat.seat_number)
          "
          @click.stop="confirmBooking(seat.seat_number)"
        >
          Confirm Booking
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.container {
  padding: 20px;
}
.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.seat-grid {
  display: grid;
  grid-template-columns: repeat(5, 120px);
  gap: 16px;
  margin-top: 20px;
}
.seat {
  padding: 20px;
  border-radius: 10px;
  color: white;
  text-align: center;
  font-weight: bold;
  cursor: pointer;
}
.available {
  background-color: green;
}
.booked {
  background-color: red;
}
.locked {
  background-color: orange;
}
</style>
