// Command gendocs menghasilkan docs/technical-defense.pdf:
// panduan jawaban technical defense untuk aplikasi MCBT.
// Jalankan ulang setiap kali isi diubah: go run ./cmd/gendocs
package main

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/go-pdf/fpdf"
)

type line struct {
	kind string // "p" paragraf, "b" bullet, "h" sub-label bold, "i" italic/kecil
	text string
}

type section struct {
	title string
	lines []line
}

var brand = struct{ r, g, b int }{61, 155, 233} // #3d9be9
var dark = struct{ r, g, b int }{30, 41, 59}    // slate-800

// Font inti PDF hanya mendukung Latin-1: transliterasi karakter UTF-8
// (panah, em-dash, box-drawing) ke ASCII agar tidak tampil sebagai "â€".
var sanitizeMap = strings.NewReplacer(
	"→", "->", "—", "-", "–", "-",
	"≥", ">=", "≤", "<=",
	"─", "-", "├", "|", "└", "+", "│", "|",
	"•", "-", "…", "...",
	"’", "'", "‘", "'", "“", "`", "”", "'",
)

// ascii memastikan string 100% Latin-1 sebelum dirender.
func ascii(s string) string {
	s = sanitizeMap.Replace(s)
	if strings.IndexFunc(s, func(r rune) bool { return r > 127 }) < 0 {
		return s
	}
	var b strings.Builder
	for _, r := range s {
		if r > 127 {
			b.WriteRune('?')
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func main() {
	genDefense()
	genStackSlide()
	genPlaybook()
}

func genDefense() {
	out := "docs/technical-defense.pdf"
	if err := os.MkdirAll("docs", 0o755); err != nil {
		log.Fatal(err)
	}

	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(16, 16, 16)
	pdf.SetAutoPageBreak(true, 18)
	pdf.SetFooterFunc(func() {
		pdf.SetY(-14)
		pdf.SetFont("Helvetica", "", 8.5)
		pdf.SetTextColor(120, 130, 145)
		pdf.CellFormat(0, 6,
			fmt.Sprintf("Panduan Technical Defense - MCBT  |  halaman %d", pdf.PageNo()),
			"", 0, "C", false, 0, "")
	})
	pdf.AddPage()

	// ---------- judul ----------
	pdf.SetTextColor(dark.r, dark.g, dark.b)
	pdf.SetFont("Helvetica", "B", 21)
	pdf.CellFormat(0, 11, "Panduan Technical Defense", "", 1, "L", false, 0, "")
	pdf.SetFont("Helvetica", "B", 14)
	pdf.SetTextColor(brand.r, brand.g, brand.b)
	pdf.CellFormat(0, 8, "Aplikasi MCBT - Computer-Based Test", "", 1, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 9.5)
	pdf.SetTextColor(110, 120, 135)
	pdf.CellFormat(0, 6, "Rangkuman keputusan teknis & alasan di baliknya - siap dibuktikan langsung di kode", "", 1, "L", false, 0, "")
	// garis brand
	pdf.SetFillColor(brand.r, brand.g, brand.b)
	pdf.Rect(16, pdf.GetY()+1, 178, 1.2, "F")
	pdf.Ln(8)

	write := func(l line) {
		l.text = ascii(l.text)
		switch l.kind {
		case "h":
			pdf.SetFont("Helvetica", "B", 10.5)
			pdf.SetTextColor(dark.r, dark.g, dark.b)
			pdf.MultiCell(0, 5.4, l.text, "", "L", false)
		case "b":
			pdf.SetFont("Helvetica", "", 10)
			pdf.SetTextColor(55, 65, 81)
			pdf.SetX(21)
			pdf.MultiCell(173, 5.2, "- "+l.text, "", "L", false)
		case "i":
			pdf.SetFont("Helvetica", "I", 9.5)
			pdf.SetTextColor(110, 120, 135)
			pdf.SetX(21)
			pdf.MultiCell(173, 5, l.text, "", "L", false)
		default:
			pdf.SetFont("Helvetica", "", 10)
			pdf.SetTextColor(55, 65, 81)
			pdf.MultiCell(0, 5.2, l.text, "", "L", false)
		}
	}

	for i, sec := range content() {
		// heading pertanyaan
		pdf.Ln(2)
		pdf.SetFillColor(240, 246, 253)
		pdf.SetFont("Helvetica", "B", 12)
		pdf.SetTextColor(23, 60, 100)
		pdf.CellFormat(0, 8.5, fmt.Sprintf("%d. %s", i+1, sec.title), "", 1, "L", true, 0, "")
		pdf.Ln(1.5)
		for _, l := range sec.lines {
			write(l)
			if l.kind == "b" {
				continue
			}
		}
		pdf.Ln(2.5)
	}

	if err := pdf.OutputFileAndClose(out); err != nil {
		log.Fatal(err)
	}
	fmt.Println("PDF dibuat:", out, fmt.Sprintf("(%d halaman)", pdf.PageNo()))
}

func content() []section {
	return []section{
		{
			title: "Kenapa memilih domain CBT / Ujian?",
			lines: []line{
				{kind: "p", text: "Ujian berbasis kertas punya banyak titik lambat dan rawan salah: cetak soal, distribusi ke ruangan, koreksi manual, lalu rekap nilai manual. Domain ini juga punya aturan bisnis nyata yang bisa dibuktikan di kode - KKM, bobot soal, negatif marking, attempt terbatas, jendela waktu ujian, ranking - sehingga bukan sekadar CRUD kosong."},
				{kind: "b", text: "Digitalisasi ujian adalah kebutuhan nyata sekolah di Indonesia (istilah domain memakai NIS, KKM, tahun ajaran ganjil/genap, wali kelas)."},
			},
		},
		{
			title: "Siapa target user?",
			lines: []line{
				{kind: "b", text: "Admin sekolah: kelola master data (siswa, guru, mapel, tahun ajaran), akun, laporan soal."},
				{kind: "b", text: "Guru: buat bank soal & ujian, review soal sebelum diujikan, nilai esai, tangani laporan soal."},
				{kind: "b", text: "Siswa: mengerjakan ujian dengan mekanisme anti-kecurangan, lihat hasil & pembahasan, laporkan soal bermasalah."},
				{kind: "i", text: "Bukti: tiga role ini dipetakan ke grup route berbeda (router.go) dan menu sidebar berbeda (AppShell.vue)."},
			},
		},
		{
			title: "Business problem yang diselesaikan",
			lines: []line{
				{kind: "b", text: "Kecurangan: timer server-authoritative, token masuk ujian, randomisasi soal/opsi per attempt, satu attempt aktif per siswa."},
				{kind: "b", text: "Koreksi manual: skoring otomatis PG/B-S/Multi/Isian (gradeObjective), esai lewat lembar grading khusus, publish hasil sekali klik."},
				{kind: "b", text: "Soal bocor/tidak valid: fitur laporkan soal oleh siswa + halaman review soal sebelum ujian (/exams/:id/review)."},
				{kind: "b", text: "Rekap & ranking manual: otomatis per ujian/kelas + ekspor Excel/PDF."},
				{kind: "b", text: "Kolaborasi guru: data-scope - guru kelola bank/ujian miliknya, bisa melihat bank kolega (read-only)."},
			},
		},
		{
			title: "Kenapa memilih arsitektur ini?",
			lines: []line{
				{kind: "p", text: "Modular monolith berlapis: handler -> usecase -> repository. Handler hanya parsing HTTP & response, usecase berisi aturan bisnis, repository hanya akses data."},
				{kind: "b", text: "Mudah dites: 24 usecase direfactor ke consumer-defined interface (interfaces.go) sehingga 94 unit test jalan tanpa DB (fakes_test.go)."},
				{kind: "b", text: "Test menemukan bug nyata: nil-pointer di questionInExam saat ujian tanpa bank soal - tertangkap sebelum produksi."},
				{kind: "b", text: "Monolith dipilih karena skala sekolah bukan ratusan ribu RPS: deployment sederhana (satu docker-compose), transaksi lintas modul memakai ACID PostgreSQL."},
			},
		},
		{
			title: "Kenapa authentication menggunakan cookie?",
			lines: []line{
				{kind: "b", text: "XSS: token di localStorage bisa dibaca JS yang disuntik; HttpOnly cookie mustahil dibaca JS. Aplikasi ini kaya rich-text (editor soal) sehingga permukaan XSS lebih besar."},
				{kind: "b", text: "Cookie otomatis terkirim setiap request - FE tidak mengelola token manual."},
				{kind: "p", text: "Trade-off-nya adalah CSRF, dan itu ditangani berlapis (lihat poin 7)."},
			},
		},
		{
			title: "Fungsi HttpOnly / Secure / SameSite",
			lines: []line{
				{kind: "b", text: "HttpOnly: cookie tidak bisa diakses document.cookie -> token aman dari pencurian via XSS."},
				{kind: "b", text: "Secure: cookie hanya dikirim lewat HTTPS -> anti penyadapan jaringan (dev false, produksi true via env COOKIE_SECURE)."},
				{kind: "b", text: "SameSite=Strict: cookie tidak dikirim pada request lintas situs -> lapisan pertama anti-CSRF."},
				{kind: "i", text: "Semua dikonfigurasi via env - tunjuk config.go dan cookies.go saat menjelaskan."},
			},
		},
		{
			title: "Cara mengatasi CSRF",
			lines: []line{
				{kind: "p", text: "Defense in depth, tiga lapis:"},
				{kind: "b", text: "1. SameSite=Strict sebagai lapisan pasif - cookie tak ikut request lintas situs."},
				{kind: "b", text: "2. Double-submit cookie: server menanam cookie csrf_token; FE (interceptor axios) wajib mengirim header X-CSRF-Token bernilai sama untuk POST/PUT/PATCH/DELETE; middleware CSRFProtection membandingkan keduanya. Attacker domain lain tidak bisa membaca cookie -> tidak bisa menyusun header."},
				{kind: "b", text: "3. Bukti nyata: dua bug dulu (lapor soal & tangani laporan memakai fetch tanpa header) ditolak 403 oleh middleware, lalu diperbaiki lewat instance axios."},
			},
		},
		{
			title: "Cara refresh authentication",
			lines: []line{
				{kind: "b", text: "Dua cookie terpisah: access_token (15 menit) dan refresh_token (7 hari)."},
				{kind: "b", text: "Interceptor FE menangkap 401 -> panggil POST /auth/refresh-token -> server validasi refresh JWT -> terbitkan access baru -> request asli di-replay. User tidak pernah logout paksa."},
				{kind: "b", text: "Rotasi: refresh juga menerbitkan pasangan token baru."},
				{kind: "b", text: "Revocation: token_version di DB; reset password menaikkan versi sehingga semua token lama otomatis ditolak (Session revoked). Dites di TestRefresh_StaleTokenVersion."},
			},
		},
		{
			title: "Cara authorization diterapkan",
			lines: []line{
				{kind: "b", text: "Role-based (RequireRoles): grup route adminOnly, staffOnly (admin+guru), candidate (siswa) di router.go."},
				{kind: "b", text: "Data-scope/ownership: bank soal & ujian punya created_by; middleware dataScopeGuard + AccessUsecase.Assert*Owner memverifikasi berantai (laporan -> attempt -> ujian -> created_by)."},
				{kind: "b", text: "Response membawa flag can_manage sehingga FE menyembunyikan tombol edit/hapus data orang lain."},
				{kind: "b", text: "Siswa hanya melihat ujian yang di-assign kepadanya dan hasil miliknya sendiri."},
			},
		},
		{
			title: "Kenapa menggunakan Composite (aggregated) API",
			lines: []line{
				{kind: "b", text: "Satu endpoint mengembalikan objek gabungan siap pakai: GET /exams menyertakan subject, question_bank, academic_year, attempts_count; dashboard mengagregasi 8 metrik dalam satu call; review soal mengembalikan soal+opsi+kunci+pembahasan per section."},
				{kind: "b", text: "Alasan: mengurangi round-trip dan state di FE (satu halaman ujian = satu request), format envelope konsisten {success, message, data}."},
				{kind: "b", text: "Trade-off: over-fetching untuk kasus tertentu - diterima karena konsumennya jelas (SPA sendiri), dan N+1 dicegah di SQL (subquery attempts_count, Preload)."},
			},
		},
		{
			title: "Cara menangani concurrent request",
			lines: []line{
				{kind: "b", text: "Jawaban siswa: unique index (attempt_id, question_id) + UPSERT -> request simultan tidak menciptakan duplikat."},
				{kind: "b", text: "Submit idempotent: attempt sudah submitted -> return tanpa menulis ulang (dites di TestAttemptSubmit/idempotent)."},
				{kind: "b", text: "Operasi multi-baris (clone bank, finalize) dibungkus transaksi database."},
				{kind: "b", text: "Sumber waktu = server: heartbeat menghitung sisa waktu dari jam server (now bisa diinjeksi & dites), jam browser tidak dipercaya."},
				{kind: "b", text: "Satu attempt aktif per (exam, student) dijaga FindActive + constraint; connection pool Postgres dikonfigurasi (DB_MAX_OPEN_CONNS)."},
			},
		},
		{
			title: "Desain database jika data menjadi sangat besar",
			lines: []line{
				{kind: "p", text: "Baris terbesar adalah exam_attempts + exam_answers (tumbuh linear per ujian). Rencananya:"},
				{kind: "b", text: "Partitioning native Postgres per exam_id atau per bulan started_at - query selalu ter-filter exam/tanggal."},
				{kind: "b", text: "Indexing sudah disiapkan sejak migrasi: composite (exam_id, student_id), idx_attempts_student, unique (attempt_id, question_id), index created_by."},
				{kind: "b", text: "Read replica untuk beban laporan/ekspor; arsip ujian lama ke tabel arsip."},
				{kind: "b", text: "Media keluar dari DB ke MinIO/S3 (hanya media_key di DB); Redis (sudah ada di compose) untuk cache ranking/dashboard."},
			},
		},
		{
			title: "Kenapa menggunakan queue?",
			lines: []line{
				{kind: "p", text: "Jawaban jujur: belum terpakai saat ini, Redis sudah disiapkan di compose - keputusan sadar (YAGNI). Queue dibutuhkan untuk pekerjaan berat fire-and-forget, kandidat nyatanya:"},
				{kind: "b", text: "Penilaian massal saat ratusan siswa submit serentak (calculate-grades jadi job)."},
				{kind: "b", text: "Generate ekspor PDF/Excel besar agar tidak memblok HTTP request."},
				{kind: "b", text: "Notifikasi (hasil rilis / pengumuman ujian). Implementasi pilihan: asynq/river di atas Redis yang sudah ada."},
			},
		},
		{
			title: "Cara menangani failed job",
			lines: []line{
				{kind: "b", text: "Retry dengan exponential backoff untuk error transient (DB timeout, S3 5xx)."},
				{kind: "b", text: "Idempotency: job penilaian aman diulang karena UpdateGrading/UpdateAttemptScore berbasis ID dan total dihitung ulang dari DB, bukan akumulasi."},
				{kind: "b", text: "Dead-letter queue setelah N kali gagal + log terstruktur (slog) dengan request_id untuk investigasi."},
				{kind: "b", text: "Partial failure: CalculateGrades memproses per attempt; esai tanpa nilai dilewati, bukan error."},
			},
		},
		{
			title: "Cara memastikan file upload aman",
			lines: []line{
				{kind: "b", text: "Ukuran dibatasi keras: MaxUploadMB dicek di usecase + client_max_body_size di Nginx + MaxMultipartMemory Gin."},
				{kind: "b", text: "Nama file asli dibuang -> disimpan dengan random key (migrasi media_key_path) -> tak ada path traversal / overwrite."},
				{kind: "b", text: "Disimpan di MinIO/S3, bukan disk aplikasi -> terpisah dari runtime, tidak bisa dieksekusi."},
				{kind: "b", text: "Disajikan lewat endpoint berotorisasi GET /media/:id/file dengan content-type terkontrol; hanya untuk gambar soal via editor rich-text oleh staff."},
			},
		},
		{
			title: "Cara meningkatkan performance aplikasi",
			lines: []line{
				{kind: "h", text: "Sudah dilakukan:"},
				{kind: "b", text: "Pagination di semua list + meta; anti-N+1 (Preload terarah, attempts_count via subquery, agregat dashboard dihitung 1 query SQL)."},
				{kind: "b", text: "Index strategis (composite & unique) sejak migrasi; SPA statik dilayani Nginx (cache + gzip), API hanya JSON ringan; skor dihitung di SQL (SUM)."},
				{kind: "h", text: "Roadmap (dengan alasan):"},
				{kind: "b", text: "Cache Redis untuk ranking/dashboard (stack sudah ada), CDN di depan MinIO untuk media, partitioning sesuai poin 12."},
			},
		},
		{
			title: "Trade-off dari arsitektur yang dipilih",
			lines: []line{
				{kind: "b", text: "Modular monolith: deploy mudah + transaksi ACID lintas modul VS scale-out per-modul terbatas; dimitigasi dengan batas lapisan yang disiplin."},
				{kind: "b", text: "JWT di cookie HttpOnly: anti-XSS VS rentan CSRF -> dibayar pajaknya dengan CSRF token + SameSite."},
				{kind: "b", text: "Server-side timer: anti-cheat kuat VS butuh heartbeat berkala; mitigasi autosave agar jawaban tidak hilang."},
				{kind: "b", text: "Composite response: FE sederhana VS over-fetch; konsumen jelas (SPA sendiri) sehingga terkendali."},
				{kind: "b", text: "GORM: produktivitas VS kontrol query - query kompleks tetap raw SQL (dashboard, ranking): pendekatan hybrid."},
				{kind: "b", text: "Interface + hand-written fakes: test cepat tanpa DB & tanpa dependensi mock VS fakes harus sinkron dengan kontrak repo ((nil,nil) vs ErrRecordNotFound)."},
				{kind: "b", text: "Redis disiapkan tanpa queue: tidak over-engineer VS fitur async menunggu kebutuhan nyata."},
			},
		},
		{
			title: "Kunci saat presentasi",
			lines: []line{
				{kind: "p", text: "Gunakan pola: keputusan -> alasan -> bukti di kode (sebut file) -> trade-off yang disadari -> mitigasinya."},
				{kind: "p", text: "Contoh kalimat pembuka: \"Saya tidak memilih teknologi karena populer, tapi karena masalah yang diselesaikan - dan setiap trade-off-nya saya akui sekaligus saya mitigasi.\""},
				{kind: "i", text: "Angka pendukung: 94 unit test usecase (coverage in-scope penuh), 5 paket Go + 7 component test FE hijau, 21 migrasi terstruktur, docker-compose 5 layanan."},
			},
		},
	}
}

// genStackSlide membuat docs/stack-overview.pdf:
// satu halaman landscape berisi diagram tech stack (siap jadi slide).
func genStackSlide() {
	out := "docs/stack-overview.pdf"
	if err := os.MkdirAll("docs", 0o755); err != nil {
		log.Fatal(err)
	}

	pdf := fpdf.New("L", "mm", "A4", "")
	pdf.SetMargins(14, 12, 14)
	pdf.SetAutoPageBreak(false, 0) // layout slide fixed - tanpa auto break
	pdf.AddPage()

	const (
		left  = 14.0
		width = 269.0
		cx    = left + width/2
	)

	brandCol := func() (int, int, int) { return brand.r, brand.g, brand.b }

	// ---- judul ----
	pdf.SetTextColor(dark.r, dark.g, dark.b)
	pdf.SetFont("Helvetica", "B", 20)
	pdf.CellFormat(0, 10, ascii("Tech Stack - MCBT (Computer-Based Test)"), "", 1, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 10)
	pdf.SetTextColor(110, 120, 135)
	pdf.CellFormat(0, 6, "Go + Gin + GORM  |  PostgreSQL  |  Vue 3 + TypeScript  |  Docker Compose", "", 1, "L", false, 0, "")
	r, g, b := brandCol()
	pdf.SetFillColor(r, g, b)
	pdf.Rect(left, pdf.GetY()+1, width, 1.1, "F")

	// ---- layer box ----
	layer := func(y, h float64, name, tech string, fr, fg, fb int) {
		pdf.SetFillColor(fr, fg, fb)
		pdf.RoundedRect(left, y, width, h, 3.5, "1234", "F")
		pdf.SetTextColor(255, 255, 255)

		pdf.SetXY(left+8, y)
		pdf.SetFont("Helvetica", "B", 13)
		pdf.CellFormat(46, h, name, "", 0, "L", false, 0, "")

		pdf.SetXY(left+56, y)
		pdf.SetFont("Helvetica", "", 11)
		pdf.CellFormat(width-64, h, ascii(tech), "", 0, "L", false, 0, "")
	}

	arrow := func(yTop, yBottom float64, label string) {
		pdf.SetFillColor(148, 163, 184)
		pdf.Polygon([]fpdf.PointType{
			{X: cx - 4, Y: yTop},
			{X: cx + 4, Y: yTop},
			{X: cx, Y: yTop + 6.5},
		}, "F")
		if label != "" {
			pdf.SetXY(cx+8, yTop-1.5)
			pdf.SetFont("Helvetica", "I", 9)
			pdf.SetTextColor(110, 120, 135)
			pdf.CellFormat(126, 5, ascii(label), "", 0, "L", false, 0, "")
		}
	}

	// ---- judul selesai di ~y26; mulai layer ----
	layer(32, 28, "CLIENT",
		"Vue 3 + TypeScript  |  Vite  |  Tailwind CSS v4  |  Pinia  |  Vue Router  |  Axios  |  Vitest",
		52, 211, 153) // hijau
	arrow(60, 69, "HTTPS  -  cookie HttpOnly + X-CSRF-Token")

	layer(69, 20, "EDGE",
		"Nginx  -  serve SPA statik (cache+gzip)  +  reverse proxy /api",
		51, 65, 85) // slate
	arrow(89, 98, "reverse proxy /api  ->  :8080")

	layer(98, 28, "API (GO)",
		"Go 1.27  |  Gin  |  GORM  |  JWT cookie + CSRF  |  bcrypt  |  excelize  |  go-pdf  |  slog",
		61, 155, 233) // biru
	arrow(126, 135, "SQL (GORM)  -  S3 API (media soal)")

	layer(135, 28, "DATA",
		"PostgreSQL 16  |  Redis 7 (cache/queue)  |  MinIO - S3 compatible (media)",
		139, 92, 246) // ungu

	// ---- strip ops & testing ----
	pdf.SetFillColor(241, 245, 249)
	pdf.RoundedRect(left, 171, width, 16, 3.5, "1234", "F")
	pdf.SetTextColor(dark.r, dark.g, dark.b)
	pdf.SetXY(left+8, 171)
	pdf.SetFont("Helvetica", "B", 10)
	pdf.CellFormat(40, 16, "OPS & TEST", "", 0, "L", false, 0, "")
	pdf.SetXY(left+50, 171)
	pdf.SetFont("Helvetica", "", 10)
	pdf.CellFormat(width-58, 16,
		ascii("Docker Compose - 5 layanan (web, api, postgres, redis, minio)  |  golang-migrate 21 versi  |  94 unit test usecase  |  7 component test FE"),
		"", 0, "L", false, 0, "")

	pdf.SetXY(left, 193)
	pdf.SetFont("Helvetica", "I", 8)
	pdf.SetTextColor(148, 163, 184)
	pdf.CellFormat(0, 5, "dibuat otomatis - go run ./cmd/gendocs", "", 0, "L", false, 0, "")

	if err := pdf.OutputFileAndClose(out); err != nil {
		log.Fatal(err)
	}
	fmt.Println("PDF dibuat:", out, fmt.Sprintf("(%d halaman)", pdf.PageNo()))
}

// genPlaybook membuat docs/rebuild-playbook.pdf dari docs/REBUILD_PLAYBOOK.md.
func genPlaybook() {
	const (
		inMd = "docs/REBUILD_PLAYBOOK.md"
		out  = "docs/rebuild-playbook.pdf"
	)
	data, err := os.ReadFile(inMd)
	if err != nil {
		log.Fatalf("baca %s: %v", inMd, err)
	}

	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(14, 14, 14)
	pdf.SetAutoPageBreak(true, 16)
	pdf.SetFooterFunc(func() {
		pdf.SetY(-13)
		pdf.SetFont("Courier", "", 8)
		pdf.SetTextColor(120, 130, 145)
		pdf.CellFormat(0, 6, fmt.Sprintf("Rebuild Playbook - MCBT  |  %d", pdf.PageNo()), "", 0, "C", false, 0, "")
	})
	pdf.AddPage()

	// judul dokumen
	pdf.SetTextColor(dark.r, dark.g, dark.b)
	pdf.SetFont("Helvetica", "B", 18)
	pdf.MultiCell(0, 9, "Rebuild Playbook - MCBT", "", "L", false)
	pdf.SetFont("Helvetica", "", 10)
	pdf.SetTextColor(110, 120, 135)
	pdf.MultiCell(0, 5.5, ascii("Panduan eksekusi membangun aplikasi dari nol: database, backend, frontend, infrastruktur, verifikasi."), "", "L", false)
	pdf.SetFillColor(brand.r, brand.g, brand.b)
	pdf.Rect(14, pdf.GetY()+1, 182, 1.1, "F")
	pdf.Ln(4)

	linkRe := regexp.MustCompile(`\[([^\]]*)\]\([^)]*\)`)

	renderCode := func(text string) {
		text = ascii(text)
		pdf.SetFont("Courier", "", 7.6)
		pdf.SetFillColor(243, 244, 246)
		pdf.SetTextColor(40, 48, 60)
		pdf.MultiCell(0, 3.4, text, "", "L", true)
		pdf.Ln(1)
	}

	inCode := false
	firstH2 := true
	for _, raw := range strings.Split(string(data), "\n") {
		line := strings.TrimRight(raw, " \t\r")

		if strings.HasPrefix(line, "```") {
			inCode = !inCode
			continue
		}
		if inCode {
			renderCode(line)
			continue
		}

		switch {
		case strings.HasPrefix(line, "# "):
			// judul utama sudah dirender manual
		case strings.HasPrefix(line, "## "):
			title := strings.TrimPrefix(line, "## ")
			if !firstH2 {
				pdf.AddPage()
			}
			firstH2 = false
			pdf.Ln(1)
			pdf.SetFont("Helvetica", "B", 14)
			pdf.SetTextColor(23, 60, 100)
			pdf.MultiCell(0, 7.5, ascii(title), "", "L", false)
			pdf.SetFillColor(brand.r, brand.g, brand.b)
			pdf.Rect(14, pdf.GetY()+0.5, 60, 0.9, "F")
			pdf.Ln(3)
		case strings.HasPrefix(line, "### "):
			title := strings.TrimPrefix(line, "### ")
			pdf.Ln(1.5)
			pdf.SetFont("Helvetica", "B", 11)
			pdf.SetTextColor(dark.r, dark.g, dark.b)
			pdf.MultiCell(0, 5.8, ascii(title), "", "L", false)
			pdf.Ln(0.5)
		case strings.HasPrefix(line, "- "):
			pdf.SetFont("Helvetica", "", 9.5)
			pdf.SetTextColor(55, 65, 81)
			pdf.SetX(19)
			pdf.MultiCell(177, 4.8, ascii("- "+linkRe.ReplaceAllString(strings.ReplaceAll(strings.TrimPrefix(line, "- "), "**", ""), "$1")), "", "L", false)
		case strings.HasPrefix(line, "> "):
			pdf.SetFont("Helvetica", "I", 9.5)
			pdf.SetTextColor(90, 100, 115)
			pdf.SetX(19)
			pdf.MultiCell(177, 4.8, ascii(linkRe.ReplaceAllString(strings.TrimPrefix(line, "> "), "$1")), "", "L", false)
		case strings.HasPrefix(line, "|"):
			if strings.ReplaceAll(line, "|-", "") == line || !strings.Contains(line, "--") {
				pdf.SetFont("Courier", "", 7.4)
				pdf.SetTextColor(40, 48, 60)
				pdf.MultiCell(0, 3.6, ascii(line), "", "L", false)
			}
		case line == "---":
			pdf.SetDrawColor(226, 232, 240)
			pdf.Line(14, pdf.GetY()+1.5, 196, pdf.GetY()+1.5)
			pdf.Ln(3)
		case line == "":
			pdf.Ln(1.6)
		default:
			pdf.SetFont("Helvetica", "", 9.5)
			pdf.SetTextColor(55, 65, 81)
			pdf.MultiCell(0, 4.8, ascii(linkRe.ReplaceAllString(strings.ReplaceAll(line, "**", ""), "$1")), "", "L", false)
		}
	}

	if err := pdf.OutputFileAndClose(out); err != nil {
		log.Fatal(err)
	}
	fmt.Println("PDF dibuat:", out, fmt.Sprintf("(%d halaman)", pdf.PageNo()))
}
