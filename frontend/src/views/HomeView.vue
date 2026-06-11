<script setup>
import { ref, onMounted } from 'vue'

const seats = ref([])

const fetchSeats = async () => {
  const response = await fetch('http://localhost:8080/seats')
  const data = await response.json()

  seats.value = data
}

onMounted(() => {
  fetchSeats()

  const socket = new WebSocket('ws://localhost:8080/ws')

  socket.onmessage = () => {
    fetchSeats()
  }
})

const getSeatClass = (status) => {
  if (status === 'AVAILABLE') return 'available'
  if (status === 'BOOKED') return 'booked'
  if (status === 'LOCKED') return 'locked'
}

const lockSeat = async (seatNumber) => {
  await fetch('http://localhost:8080/lock-seat', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      seat_number: seatNumber
    })
  })

  fetchSeats()
}
const confirmBooking = async (seatNumber) => {

  const response = await fetch('http://localhost:8080/confirm-booking', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      user_id: 'user123',
      seat_number: seatNumber
    })
  })

  const data = await response.json()

  alert(data.message || data.error)

  fetchSeats()
}
</script>

<template>
  <div class="container">
    <h1>Cinema Seat Map</h1>

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
          v-if="seat.status === 'LOCKED'"
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