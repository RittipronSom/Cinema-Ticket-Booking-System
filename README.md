# Cinema Ticket Booking System

ระบบจองตั๋วภาพยนตร์ออนไลน์ที่รองรับการจองพร้อมกันหลายคนโดยไม่เกิด Double Booking

---

## 1. System Architecture

```
┌─────────────────┐         ┌──────────────────────────────────────┐
│   Frontend      │         │            Backend (Go + Gin)        │
│   Vue 3         │◄───────►│                                      │
│   Firebase Auth │  HTTP   │  /seats        GET  (public)         │
│   WebSocket     │  WS     │  /lock-seat    POST (auth required)  │
                            │  /unlock-seat  POST (auth required)  │
└─────────────────┘         │  /confirm      POST (auth required)  │
                            │  /bookings     GET  (admin only)     │
                            │  /ws           WS   (public)         │
                            └──────┬───────────────┬──────────────┘
                                   │               │
                          ┌────────▼───────┐  ┌────▼──────────┐
                          │   MongoDB      │  │   Redis        │
                          │  - bookings    │  │  - seat locks  │
                          │  - audit_logs  │  │  - pub/sub     │
                          └────────────────┘  └───────────────┘
```

---

## 2. Tech Stack

| Layer | Technology |
|---|---|
| Backend | Go + Gin |
| Frontend | Vue 3 |
| Database | MongoDB 7 |
| Cache / Lock | Redis 7 |
| Realtime | WebSocket |
| Message Queue | Redis Pub-Sub |
| Auth | Firebase Authentication (Google OAuth) |
| Deployment | Docker + docker-compose |

---

## 3. Booking Flow

```
Step 1: ผู้ใช้ Login ด้วย Google (Firebase Auth)
           ↓
Step 2: ผู้ใช้เลือกที่นั่งที่ต้องการ (AVAILABLE)
           ↓
Step 3: Frontend ส่ง POST /lock-seat พร้อม Firebase Token
           ↓
Step 4: Backend verify Token → ใช้ Redis SetNX lock ที่นั่ง 5 นาที
         - ถ้า lock สำเร็จ → สถานะเปลี่ยนเป็น LOCKED
         - ถ้า lock ไม่สำเร็จ → ที่นั่งถูกจองแล้ว return error
           ↓
Step 5: WebSocket broadcast → ผู้ใช้คนอื่นเห็นสถานะ LOCKED ทันที
           ↓
Step 5.5: ผู้ใช้กดที่นั่ง LOCKED ที่ตัวเองล็อคไว้ → ยกเลิก lock ทันที (Unlock)
           ↓
Step 6: ผู้ใช้กด Confirm (ภายใน 5 นาที) → จองทุกที่นั่งที่ล็อคไว้พร้อมกัน
           ↓
Step 7: Backend บันทึกลง MongoDB → สถานะเปลี่ยนเป็น BOOKED
         → Publish event ไปยัง Redis channel "booking_events"
         → ลบ Redis lock key
           ↓
Step 8: WebSocket broadcast → ทุกคนเห็นสถานะ BOOKED ทันที

--- กรณีไม่ชำระเงิน ---

Step 6': ครบ 5 นาที → Redis key หมดอายุอัตโนมัติ
           ↓
Step 7': LockExpiryWatcher ตรวจพบภายใน 3 วินาที
           ↓
Step 8': สถานะเปลี่ยนกลับเป็น AVAILABLE + WebSocket broadcast
```

---

## 4. Redis Lock Strategy

ใช้ **Redis Distributed Lock** ด้วยคำสั่ง `SetNX` (Set if Not Exists)

```
Key:   "seat:A1"
Value: "locked"
TTL:   5 นาที
```

**ทำไมถึงเลือก SetNX:**
- Atomic operation — ป้องกัน Race Condition เมื่อมีหลายคนกดพร้อมกัน
- TTL อัตโนมัติ — ถ้าผู้ใช้ไม่ยืนยันภายใน 5 นาที ระบบคืน lock ให้คนอื่นได้เลย
- ไม่ต้อง cleanup เอง — Redis จัดการหมดอายุให้

**Lock Expiry Watcher:**
Background goroutine ตรวจสอบทุก 3 วินาที ถ้า Redis key หายไปแต่ที่นั่งยังเป็น LOCKED จะเปลี่ยนกลับเป็น AVAILABLE และ broadcast ผ่าน WebSocket ทันที

---

## 5. Message Queue (Redis Pub-Sub)

ใช้ Redis Pub-Sub สำหรับ Async Event Processing

**Channel:** `booking_events`

**Use Case:**
- เมื่อ Booking สำเร็จ → Publish event ไปที่ channel
- Subscriber รับ event → Mock Notification (สามารถต่อยอดเป็น Email/LINE ได้)

**ทำไมถึงเลือก Redis Pub-Sub แทน Kafka:**
- ใช้ Redis ที่มีอยู่แล้วได้เลย ไม่ต้องติดตั้งเพิ่ม
- เหมาะกับ use case ขนาดนี้ที่ไม่ต้องการ message persistence

---

## 6. วิธีรันระบบ

**Requirements:**
- Docker
- Docker Compose

### Setup ก่อนรัน (สำคัญ)

**Step 1 — สร้าง Firebase Project**
1. ไปที่ [Firebase Console](https://console.firebase.google.com) → สร้าง project ใหม่
2. เปิด **Authentication** → **Sign-in method** → เปิด **Google**
3. คลิก **Project Settings** (ไอคอนเฟือง) → แท็บ **Service accounts**
4. กด **Generate new private key** → ดาวน์โหลดไฟล์ `.json`
5. **เปลี่ยนชื่อไฟล์เป็น `serviceAccountKey.json`** แล้ววางไว้ใน folder `backend/`
6. ไปที่ **Project Settings** → **General** → **Your apps** → เพิ่ม Web app
7. copy `firebaseConfig` แล้วแทนที่ใน `frontend/src/firebase.js`

**Step 2 — สร้างไฟล์ `backend/.env`**
```
ADMIN_SECRET=cinema-admin-2024
MONGO_URI=mongodb://mongo:27017
REDIS_ADDR=redis:6379
```

**Step 3 — รันระบบ**
```bash
docker compose up --build
```

**URL:**
- Frontend: `http://localhost:5173`
- Backend API: `http://localhost:8080`
- Admin Dashboard: `http://localhost:5173/admin`

---

### ทดสอบ Admin Dashboard

หน้า Admin ต้องการ Google Login ที่มี email ตรงกับที่กำหนดใน `frontend/src/router/index.js` และ `frontend/src/views/HomeView.vue`

เพิ่ม email ของคุณใน 2 ไฟล์นี้:

**`frontend/src/router/index.js`:**
```javascript
const ADMIN_EMAILS = ['your-email@gmail.com']
```

**`frontend/src/views/HomeView.vue`:**
```vue
v-if="user.email === 'your-email@gmail.com'"
```

หรือทดสอบ Admin API โดยตรงผ่าน curl:
```bash
curl http://localhost:8080/bookings -H "X-Admin-Secret: cinema-admin-2024"
```

---

## 7. Assumptions & Trade-offs

**Assumptions:**
- ที่นั่งถูก hardcode ไว้ 4 ที่ (A1-A4) เพื่อความง่ายในการ demo
- Admin ถูก hardcode ด้วย email ใน frontend และ secret key ใน backend
- ไม่มี Payment Gateway จริง กด Confirm = ชำระเงินสำเร็จ

**Trade-offs:**
- `seats` slice เก็บใน memory → เร็วแต่ถ้ามีหลาย instance จะไม่ sync กัน ทางแก้คือใช้ Redis เก็บ state แทน
- LockExpiryWatcher ใช้ polling ทุก 3 วินาที → ง่ายแต่ไม่ efficient เท่า Redis Keyspace Notifications
- Admin secret อยู่ใน frontend → ควรย้ายไปใช้ JWT + Role claim แทนใน production

---

## 8. การเคลียข้อมูล

**ลบข้อมูลการจองทั้งหมด:**

```bash
docker exec -it cinema-ticket-booking-system-mongo-1 mongosh
use cinema_booking
db.bookings.deleteMany({})
db.audit_logs.deleteMany({})
```

**ลบ Redis lock ที่ค้างอยู่:**

```bash
docker exec -it cinema-ticket-booking-system-redis-1 redis-cli FLUSHALL
```
