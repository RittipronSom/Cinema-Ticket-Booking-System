<script setup>
import { onMounted, ref, computed } from "vue";

const bookings = ref([]);
const seatFilter = ref("");

const fetchBookings = async () => {
  const response = await fetch("http://localhost:8080/bookings");

  const data = await response.json();

  bookings.value = data;
};

const filteredBookings = computed(() => {
  if (!seatFilter.value) {
    return bookings.value;
  }

  return bookings.value.filter((booking) =>
    booking.seat_number.toLowerCase().includes(seatFilter.value.toLowerCase()),
  );
});

onMounted(() => {
  fetchBookings();
});
</script>

<template>
  <div class="container">
    <h1>Admin Dashboard</h1>
    <input v-model="seatFilter" placeholder="Filter by seat" />

    <table border="1" cellpadding="10">
      <thead>
        <tr>
          <th>User</th>
          <th>Seat</th>
          <th>Status</th>
        </tr>
      </thead>

      <tbody>
        <tr v-for="booking in filteredBookings" :key="booking.id">
          <td>{{ booking.user_id }}</td>
          <td>{{ booking.seat_number }}</td>
          <td>{{ booking.status }}</td>
        </tr>
      </tbody>
    </table>
  </div>
</template>
