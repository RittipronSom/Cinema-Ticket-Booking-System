<script setup>
import { onMounted, ref, computed } from "vue";

const bookings = ref([]);
const seatFilter = ref("");

const fetchBookings = async () => {
  const response = await fetch("http://localhost:8080/bookings", {
    headers: {
      "X-Admin-Secret": "cinema-admin-2024"
    }
  })
  const data = await response.json()
  bookings.value = data
}

const filteredBookings = computed(() => {
  if (!seatFilter.value) {
    return bookings.value;
  }

  return bookings.value.filter((booking) =>
    booking.seat_number.toLowerCase().includes(seatFilter.value.toLowerCase()),
  );
});

onMounted(() => {
  console.log('AdminView mounted');
  fetchBookings();
});
</script>

<template>
  <div class="page">
    <header class="header">
      <div class="brand">CINEMA <span class="divider">/</span> <span class="brand-sub">ADMIN</span></div>
      <button class="btn btn-ghost" @click="$router.push('/')">Back</button>
    </header>

    <main class="main">
      <div class="intro">
        <span class="eyebrow">Dashboard</span>
        <h1 class="title">Bookings</h1>
        <p class="subtitle">All reservations in the system.</p>
      </div>

      <div class="toolbar">
        <input
          v-model="seatFilter"
          class="search"
          placeholder="Filter by seat..."
        />
        <span class="count">{{ filteredBookings.length }} results</span>
      </div>

      <div class="table-wrap">
        <table class="table">
          <thead>
            <tr>
              <th>User</th>
              <th>Seat</th>
              <th>Status</th>
            </tr>
          </thead>

          <tbody>
            <tr v-for="booking in filteredBookings" :key="booking.id">
              <td class="user">{{ booking.user_id }}</td>
              <td class="seat-cell">{{ booking.seat_number }}</td>
              <td>
                <span class="badge">{{ booking.status }}</span>
              </td>
            </tr>
            <tr v-if="filteredBookings.length === 0">
              <td colspan="3" class="empty">No bookings found.</td>
            </tr>
          </tbody>
        </table>
      </div>
    </main>
  </div>
</template>

<style scoped>
.page {
  min-height: 100vh;
  background: var(--bg);
  color: var(--fg);
}

.header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 28px 64px;
  border-bottom: 1px solid var(--border);
  background: var(--bg);
}

.brand {
  font-size: 22px;
  font-weight: 800;
  letter-spacing: 0.2em;
  color: var(--blue);
  display: flex;
  align-items: center;
  gap: 8px;
}

.brand-dot {
  color: var(--blue);
}

.divider {
  color: var(--fg-soft);
  font-weight: 400;
  margin: 0 4px;
}

.brand-sub {
  font-size: 16px;
  font-weight: 600;
  letter-spacing: 0.2em;
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

.main {
  max-width: 1000px;
  margin: 0 auto;
  padding: 48px 32px;
}

.intro {
  text-align: center;
  margin-bottom: 32px;
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

.toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 28px;
  gap: 16px;
}

.search {
  flex: 1;
  max-width: 360px;
  padding: 12px 20px;
  font-size: 16px;
  border: 1px solid var(--border);
  border-radius: 999px;
  background: var(--bg);
  color: var(--fg);
  outline: none;
  transition: all 0.2s ease;
}

.search:focus {
  border-color: var(--blue);
  box-shadow: 0 0 0 3px var(--blue-pale);
}

.search::placeholder {
  color: var(--fg-soft);
}

.count {
  font-size: 13px;
  font-weight: 600;
  color: var(--fg-muted);
  letter-spacing: 0.1em;
  text-transform: uppercase;
}

.table-wrap {
  border: 1px solid var(--border);
  border-radius: 16px;
  overflow: hidden;
  background: var(--bg);
  box-shadow: 0 4px 24px rgba(26, 78, 216, 0.04);
}

.table {
  width: 100%;
  border-collapse: collapse;
  font-size: 16px;
}

.table thead th {
  text-align: left;
  font-weight: 700;
  font-size: 12px;
  letter-spacing: 0.2em;
  text-transform: uppercase;
  color: var(--blue);
  padding: 20px 28px;
  background: var(--blue-pale);
  border-bottom: 1px solid var(--border);
}

.table tbody td {
  padding: 22px 28px;
  border-bottom: 1px solid var(--border);
}

.table tbody tr:last-child td {
  border-bottom: none;
}

.table tbody tr:hover {
  background: var(--blue-pale);
}

.user {
  font-weight: 500;
  color: var(--fg);
}

.seat-cell {
  font-weight: 700;
  letter-spacing: 0.02em;
  color: var(--blue);
  font-size: 18px;
}

.badge {
  display: inline-block;
  padding: 6px 14px;
  border: 1.5px solid var(--blue);
  border-radius: 999px;
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.1em;
  color: var(--blue);
  background: var(--blue-pale);
  text-transform: uppercase;
}

.empty {
  text-align: center;
  color: var(--fg-muted);
  padding: 56px 24px;
  font-size: 16px;
}

@media (max-width: 640px) {
  .header {
    padding: 20px 24px;
  }
  .brand {
    font-size: 18px;
  }
  .brand-sub {
    font-size: 14px;
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
  .table thead th,
  .table tbody td {
    padding: 14px 16px;
  }
  .table {
    font-size: 14px;
  }
  .seat-cell {
    font-size: 16px;
  }
}
</style>
