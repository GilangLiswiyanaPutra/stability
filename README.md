# Stability Team Technical Test — Task Manager API

## Endpoint yang Tersedia

| Method | Endpoint | Keterangan |
|--------|----------|------------|
| GET | `/tasks` | Ambil semua task |
| GET | `/tasks/:id` | Ambil task berdasarkan ID |
| POST | `/tasks` | Buat task baru |
| PUT | `/tasks/:id` | Update task yang sudah ada |
| DELETE | `/tasks/:id` | Hapus task berdasarkan ID |

---

## Bug yang Ditemukan

### Bug 1 — HTTP Status Code Salah Saat Task Tidak Ditemukan

**File:** `handlers/task_handler.go` — fungsi `GetTask`

**Masalah:** Ketika task tidak ditemukan, API mengembalikan HTTP `200 OK` padahal seharusnya `404 Not Found`. Ini menyesatkan client karena status sukses dikembalikan meskipun resource tidak ada.

**Dampak:** Client (frontend atau API consumer) yang bergantung pada HTTP status code untuk menentukan alur logika akan salah menginterpretasikan kasus "tidak ditemukan" sebagai response sukses, yang dapat menyebabkan bug tersembunyi di sisi client.

---

### Bug 2 — Error dari `strconv.Atoi` Diabaikan

**File:** `handlers/task_handler.go` — fungsi `GetTask` & `DeleteTask`

**Masalah:** Error yang dikembalikan oleh `strconv.Atoi` dibuang menggunakan `_`. Jika parameter `:id` bukan angka valid (contoh: `/tasks/abc`), nilai `id` diam-diam menjadi `0` dan operasi dilanjutkan tanpa memberikan feedback error kepada client.

**Dampak:** Request dengan format ID tidak valid (contoh: `/tasks/abc`) tidak menghasilkan error. Pada `DeleteTask`, ini bisa diam-diam mencoba menghapus task dengan ID `0`. Kegagalan diam-diam seperti ini sangat sulit untuk di-debug di lingkungan produksi.

---

### Bug 3 — Pointer ke Loop Variable di `GetTaskByID`

**File:** `store/task_store.go` — fungsi `GetTaskByID`

**Masalah:** Dalam loop `for _, t := range Tasks`, variabel `t` adalah salinan (copy) dari elemen, bukan referensi ke elemen aslinya. Mengembalikan `&t` berarti mengembalikan pointer ke variabel loop sementara yang berubah di setiap iterasi, bukan ke elemen asli di slice.

**Dampak:** Data yang dikembalikan oleh `GetTaskByID` bisa tidak konsisten atau menunjuk ke data yang salah, terutama jika pointer tersebut digunakan untuk operasi lebih lanjut seperti update. Ini adalah bug klasik di Go yang bisa tampak bekerja normal dalam kondisi tertentu, namun gagal secara tidak terduga di produksi.

---

### Bug 4 — Tidak Ada Validasi Input dan Auto-Generate ID di `CreateTask`

**File:** `handlers/task_handler.go` — fungsi `CreateTask`

**Masalah:**
- Field `Title` tidak divalidasi — title kosong atau hanya spasi bisa tersimpan
- `ID` tidak di-auto-generate — client harus menyediakan ID sendiri, berpotensi menyebabkan duplikat atau ID bernilai nol
- `return err` saat `BodyParser` gagal mengembalikan raw Go error, bukan HTTP response yang proper

**Dampak:**
- Task dengan title kosong mengotori in-memory store dengan data tidak valid
- ID duplikat menyebabkan `GetTaskByID` selalu mengembalikan task pertama yang cocok, sehingga task lain tidak dapat diakses
- Mengembalikan raw Go error bisa membocorkan informasi internal server ke client, menimbulkan risiko keamanan

---

### Bug 5 — `DeleteTask` Selalu Mengembalikan Sukses

**File:** `handlers/task_handler.go` — fungsi `DeleteTask`

**Masalah:** Endpoint selalu mengembalikan `{"message": "deleted"}` tanpa peduli apakah ID yang diberikan benar-benar ada di data store.

**Dampak:** Client akan selalu mengira operasi delete berhasil, meskipun tidak ada data yang benar-benar dihapus. Ini menyebabkan inkonsistensi data antara client dan server, yang sangat berbahaya pada sistem yang menangani penghapusan data kritis seperti akun pengguna atau transaksi.

---

## Cara Memperbaikinya

### Fix 1 — Perbaiki HTTP Status Code

Mengubah status response dari `200` menjadi `404` ketika task tidak ditemukan.

```go
// Sebelum
return c.Status(200).JSON(fiber.Map{"error": "task not found"})

// Sesudah
return c.Status(404).JSON(fiber.Map{"error": "task not found"})
```

---

### Fix 2 — Tangkap Error dari `strconv.Atoi`

Menangkap dan menangani error konversi di kedua fungsi `GetTask` dan `DeleteTask`.

```go
// Sebelum
id, _ := strconv.Atoi(idParam)

// Sesudah
id, err := strconv.Atoi(idParam)
if err != nil {
    return c.Status(400).JSON(fiber.Map{"error": "invalid id, must be a number"})
}
```

---

### Fix 3 — Gunakan Index, Bukan Loop Variable

Mengganti pointer ke loop variable dengan akses berbasis index agar pointer mengarah ke elemen asli di slice.

```go
// Sebelum
for _, t := range Tasks {
    if t.ID == id {
        return &t
    }
}

// Sesudah
for i := range Tasks {
    if Tasks[i].ID == id {
        return &Tasks[i]
    }
}
```

---

### Fix 4 — Tambah Validasi Input dan Auto-Generate ID

Menambahkan penanganan error body request yang proper, validasi title, dan auto-generate ID melalui fungsi counter `NextID()` di store.

```go
// Sesudah
if err := c.BodyParser(&task); err != nil {
    return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
}
if task.Title == "" {
    return c.Status(400).JSON(fiber.Map{"error": "title is required"})
}
task.ID = store.NextID()
store.AddTask(task)
return c.Status(201).JSON(task)
```

---

### Fix 5 — Cek Apakah Task Ada Sebelum Mengembalikan Sukses Delete

Mengubah fungsi `DeleteTask` di store agar mengembalikan `bool` yang menunjukkan apakah penghapusan berhasil, kemudian hasilnya ditangani di handler.

```go
// Store
func DeleteTask(id int) bool {
    for i, t := range Tasks {
        if t.ID == id {
            Tasks = append(Tasks[:i], Tasks[i+1:]...)
            return true
        }
    }
    return false
}

// Handler
deleted := store.DeleteTask(id)
if !deleted {
    return c.Status(404).JSON(fiber.Map{"error": "task not found"})
}
return c.JSON(fiber.Map{"message": "task deleted successfully"})
```

---

## Improvement yang Ditambahkan

### 1. Validasi Input yang Lebih Lengkap

Selain memperbaiki bug title kosong, aturan validasi tambahan ditambahkan:

- **Trim whitespace:** Title yang hanya berisi spasi sekarang ditolak menggunakan `strings.TrimSpace`
- **Batas panjang maksimal:** Title yang melebihi 100 karakter ditolak dengan pesan error yang deskriptif

```go
task.Title = strings.TrimSpace(task.Title)
if task.Title == "" {
    return errorResponse(c, 400, "title is required")
}
if len(task.Title) > 100 {
    return errorResponse(c, 400, "title must not exceed 100 characters")
}
```

---

### 2. Penanganan Error Terpusat via Response Helper

File baru `handlers/response.go` dibuat untuk memusatkan logika response dan memastikan konsistensi di semua handler.

```go
func errorResponse(c *fiber.Ctx, status int, message string) error {
    return c.Status(status).JSON(fiber.Map{"error": message})
}

func createdResponse(c *fiber.Ctx, data interface{}) error {
    return c.Status(201).JSON(data)
}

func parseID(c *fiber.Ctx) (int, error) {
    return strconv.Atoi(c.Params("id"))
}
```

Pendekatan ini menghilangkan blok `c.Status(...).JSON(fiber.Map{...})` yang berulang di setiap handler dan membuat perubahan format error di masa depan cukup dilakukan di satu tempat.

---

### 3. Endpoint Baru — `PUT /tasks/:id` (Update Task)

Endpoint baru ditambahkan untuk mendukung pembaruan `title` dan status `done` dari task yang sudah ada.

**Route:** `PUT /tasks/:id`

**Request Body:**
```json
{
  "title": "Judul yang Diperbarui",
  "done": true
}
```

**Response:**
- `200 OK` — task berhasil diperbarui, mengembalikan objek task terbaru
- `400 Bad Request` — format ID tidak valid atau validasi gagal
- `404 Not Found` — task dengan ID yang diberikan tidak ditemukan

Fungsi `UpdateTask` ditambahkan di store:
```go
func UpdateTask(id int, updated models.Task) bool {
    for i := range Tasks {
        if Tasks[i].ID == id {
            Tasks[i].Title = updated.Title
            Tasks[i].Done = updated.Done
            return true
        }
    }
    return false
}
```

---

### 4. Refactoring Kode

Semua handler di-refactor untuk menggunakan helper function terpusat, yang secara signifikan mengurangi duplikasi kode. Helper `parseID` mengekstrak dan memvalidasi parameter URL dalam satu panggilan yang dapat digunakan kembali, dan semua response error maupun sukses kini mengikuti struktur yang konsisten di seluruh codebase.

```go
// Sebelum refactor
idParam := c.Params("id")
id, err := strconv.Atoi(idParam)
if err != nil {
    return c.Status(400).JSON(fiber.Map{"error": "invalid id, must be a number"})
}

// Sesudah refactor
id, err := parseID(c)
if err != nil {
    return errorResponse(c, 400, "invalid id, must be a number")
}
```