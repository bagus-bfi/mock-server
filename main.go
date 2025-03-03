package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

var counter = 0

func main() {
	// get homepage display log request from all api
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// display log request
		// open file log.txt
		f, err := os.Open("log.txt")
		if err != nil {
			fmt.Println("Error opening file:", err)
			return
		}
		defer f.Close()

		// read file log.txt
		buf := make([]byte, 1024)
		w.Header().Set("ngrok-skip-browser-warning", "true")
		for {
			n, _ := f.Read(buf)
			if n == 0 {
				break
			}
			w.Write(buf[:n])
		}
	})

	// get token api
	http.HandleFunc("/api/v1/getToken", func(w http.ResponseWriter, r *http.Request) {
		writeLog(r)
		jsonResp := `{"code":"200","status":"success","message":"ok","data":{"valid_data":"20240830092130","token":"sometoken"}}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(jsonResp))
	})

	// post search api
	http.HandleFunc("/api/v1/product/search", func(w http.ResponseWriter, r *http.Request) {
		writeLog(r)
		// get response from file search_response.json
		jsonResp, err := os.ReadFile("search_response.json")
		if err != nil {
			fmt.Println("Error reading file:", err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(jsonResp)
	})

	// update search result api
	http.HandleFunc("/api/v1/product/updateSearchResult", func(w http.ResponseWriter, r *http.Request) {
		bodyReq, err := io.ReadAll(r.Body)
		if err != nil {
			fmt.Println("Error reading body:", err)
			w.Write([]byte("Error reading body"))
			return
		}

		// update search_response.json with request body
		err = os.WriteFile("search_response.json", bodyReq, 0644)
		if err != nil {
			fmt.Println("Error writing file:", err)
			w.Write([]byte("Error writing file"))
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Search result updated"))
	})

	// post generate report api
	http.HandleFunc("/api/v1/product/generateReport", func(w http.ResponseWriter, r *http.Request) {
		writeLog(r)
		jsonResp := `{
			"code": "01",
			"status": "Success",
			"message": "Proses membuat report sedang dikerjakan",
			"event_id": "efec27a9-yc2a-49e5-b479-78431bd78e62"
		}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(jsonResp))
	})

	// get report api
	http.HandleFunc("/api/v1/product/getReport/eventId/{id}", func(w http.ResponseWriter, r *http.Request) {
		writeLog(r)
		jsonResp := getReport()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(jsonResp))
	})

	// download report api
	http.HandleFunc("/api/v1/product/downloadReport/eventId/{id}", func(w http.ResponseWriter, r *http.Request) {
		writeLog(r)
		jsonResp := getReport()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(jsonResp))
	})

	// get pdf report api
	http.HandleFunc("/api/v1/product/downloadPdfReport/eventId/{id}", func(w http.ResponseWriter, r *http.Request) {
		writeLog(r)
		// Open the PDF file
		filePath := "NewReport.pdf"
		file, err := os.Open(filePath)
		if err != nil {
			http.Error(w, "File not found.", http.StatusNotFound)
			return
		}
		defer file.Close()

		// Set headers
		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("Content-Disposition", "attachment; filename=report.pdf")

		// Serve the file
		http.ServeFile(w, r, filePath)
	})

	// api clear log file
	http.HandleFunc("/clear", func(w http.ResponseWriter, r *http.Request) {
		// clear log file
		f, err := os.OpenFile("log.txt", os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Println("Error opening file:", err)
			return
		}

		defer f.Close()

		w.Write([]byte("Log file cleared"))
	})

	// api update pefindo_ids
	http.HandleFunc("/updatePefindoIDs", func(w http.ResponseWriter, r *http.Request) {
		// update pefindo_ids
		bodyReq, err := io.ReadAll(r.Body)
		if err != nil {
			fmt.Println("Error reading body:", err)
			w.Write([]byte("Error reading body"))
			return
		}

		type pefindoIDs struct {
			PefindoIDs []string `json:"pefindo_ids"`
		}

		var pIDs pefindoIDs
		err = json.Unmarshal(bodyReq, &pIDs)
		if err != nil {
			fmt.Println("Error unmarshalling:", err)
			w.Write([]byte("Error unmarshalling"))
			return
		}

		// write pefindo_ids to file pefindo_ids.txt
		newContent := []byte(strings.Join(pIDs.PefindoIDs, ","))
		if err := os.WriteFile("pefindo_ids.txt", newContent, 0644); err != nil {
			fmt.Println("Error writing to file:", err)
			w.Write([]byte("Error writing to file"))
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Pefindo IDs updated"))
	})

	// api run script deploy.sh
	http.HandleFunc("/deploy", func(w http.ResponseWriter, r *http.Request) {
		// run script deploy.sh
		cmd := exec.Command("sh", "deploy.sh")
		err := cmd.Run()
		if err != nil {
			fmt.Println("Error running script:", err)
			w.Write([]byte("Error running script"))
			return
		}

		// tail system log to response
		cmd = exec.Command("tail", "/var/log/syslog")
		out, err := cmd.Output()
		if err != nil {
			fmt.Println("Error running tail:", err)
			w.Write([]byte("Error running tail"))
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		w.Write(out)
	})

	log.Println("Server is running at http://localhost:9090")
	if err := http.ListenAndServe("0.0.0.0:9090", nil); err != nil {
		fmt.Println("Error starting server:", err)
	}

}

func writeLog(r *http.Request) {
	// write request log to file
	f, err := os.ReadFile("log.txt")
	if err != nil && !os.IsNotExist(err) {
		fmt.Println("Error reading file:", err)
		return
	}

	val := fmt.Sprintf("%s %s %s\n", time.Now().Format(time.DateTime), r.Method, r.URL)
	// body request
	body, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Println("Error reading body:", err)
		return
	}
	val += string(body) + "\n"

	newContent := append([]byte(val), f...)

	if err := os.WriteFile("log.txt", newContent, 0644); err != nil {
		fmt.Println("Error writing to file:", err)
	}
}

func getReport() string {
	pefindoID := getRoundRobinPefindoIDs()
	jsonResp := `{
		"code": "01",
		"status": "Success",
		"message": "Laporan berhasil di buat",
		"event_id": "efec27a9-yc2a-49e5-b479-78431bd78e60",
		"report": {
			"debitur": {
				"alamat_debitur": "Jalan Dummy 0101010290023490",
				"alamat_tempat_bekerja": null,
				"bank_jml_aktif_fasilitas": 1,
				"bank_jml_fasilitas_tertunggak": 0,
				"bank_jml_hari_tunggakan": 0,
				"bank_jml_plafon": 10900000,
				"bank_jml_saldo_terutang": 10900000,
				"bank_jml_tunggakan": 0,
				"bank_jml_tutup_fasilitas": 0,
				"bank_kolektabilitas_terburuk": 1,
				"bank_tgl_buka_fasilitas_terakhir": "2024-01-31",
				"bank_tgl_tutup_fasilitas_terakhir": null,
				"coorp_jml_aktif_fasilitas": 0,
				"coorp_jml_fasilitas_tertunggak": 0,
				"coorp_jml_hari_tunggakan": 0,
				"coorp_jml_plafon": 0,
				"coorp_jml_saldo_terutang": 0,
				"coorp_jml_tunggakan": 0,
				"coorp_jml_tutup_fasilitas": 0,
				"coorp_kolektabilitas_terburuk": 0,
				"coorp_tgl_buka_fasilitas_terakhir": null,
				"coorp_tgl_tutup_fasilitas_terakhir": null,
				"email": "ChuumonSepuluhTiga@email.com",
				"id_debitur_golden_record": 101010290023490,
				"id_golongan_debitur": "S14",
				"id_jenis_kelamin": "L",
				"id_kabupaten_kota": "0000",
				"id_lokasi": "12940",
				"id_negara": "ID",
				"id_pekerjaan": "013",
				"id_sektor_ekonomi": "X-011220",
				"id_status_gelar": "05",
				"id_status_perkawinan": "1",
				"id_tipe_debitur": 1,
				"jml_agunan": 1,
				"jml_aktif_fasilitas": 1,
				"jml_fasilitas_surat_berharga": 0,
				"jml_fasilitas_tertunggak": 0,
				"jml_hari_tunggakan": 0,
				"jml_hari_tunggakan_terburuk": 0,
				"jml_nilai_agunan": 13836540,
				"jml_penjamin": 0,
				"jml_plafon": 10900000,
				"jml_saldo_terutang": 10900000,
				"jml_tunggakan": 0,
				"jml_tutup_fasilitas": 0,
				"kecamatan": "Setiabudi",
				"kelurahan": "Setiabudi",
				"kolektabilitas_terburuk": 1,
				"mf_jml_aktif_fasilitas": 0,
				"mf_jml_fasilitas_tertunggak": 0,
				"mf_jml_hari_tunggakan": 0,
				"mf_jml_plafon": 0,
				"mf_jml_saldo_terutang": 0,
				"mf_jml_tunggakan": 0,
				"mf_jml_tutup_fasilitas": 0,
				"mf_kolektabilitas_terburuk": 0,
				"mf_tgl_buka_fasilitas_terakhir": null,
				"mf_tgl_tutup_fasilitas_terakhir": null,
				"nama_gadis_ibu_kandung": "Mom Chuumon Sepuluh Tiga",
				"nama_lengkap_debitur": "Chuumon Sepuluh Tiga",
				"nama_sesuai_identitas": "Chuumon Sepuluh Tiga",
				"nbfi_jml_aktif_fasilitas": 0,
				"nbfi_jml_fasilitas_tertunggak": 0,
				"nbfi_jml_hari_tunggakan": 0,
				"nbfi_jml_plafon": 0,
				"nbfi_jml_saldo_terutang": 0,
				"nbfi_jml_tunggakan": 0,
				"nbfi_jml_tutup_fasilitas": 0,
				"nbfi_kolektabilitas_terburuk": 0,
				"nbfi_tgl_buka_fasilitas_terakhir": null,
				"nbfi_tgl_tutup_fasilitas_terakhir": null,
				"nomor_identitas": "3175010102903493",
				"npwp": "3175010102903493",
				"p2p_jml_aktif_fasilitas": 0,
				"p2p_jml_fasilitas_tertunggak": 0,
				"p2p_jml_hari_tunggakan": 0,
				"p2p_jml_plafon": 0,
				"p2p_jml_saldo_terutang": 0,
				"p2p_jml_tunggakan": 0,
				"p2p_jml_tutup_fasilitas": 0,
				"p2p_kolektabilitas_terburuk": 0,
				"p2p_tgl_buka_fasilitas_terakhir": null,
				"p2p_tgl_tutup_fasilitas_terakhir": null,
				"pawn_jml_aktif_fasilitas": 0,
				"pawn_jml_fasilitas_tertunggak": 0,
				"pawn_jml_hari_tunggakan": 0,
				"pawn_jml_plafon": 0,
				"pawn_jml_saldo_terutang": 0,
				"pawn_jml_tunggakan": 0,
				"pawn_jml_tutup_fasilitas": 0,
				"pawn_kolektabilitas_terburuk": 0,
				"pawn_tgl_buka_fasilitas_terakhir": null,
				"pawn_tgl_tutup_fasilitas_terakhir": null,
				"tanggal_lahir": "1990-02-01",
				"telepon": "622153023593",
				"telepon_seluler": "6287821823493",
				"tempat_bekerja": null,
				"tempat_lahir": "Bandar Lampung",
				"tgl_buka_fasilitas_terakhir": "2024-01-31",
				"tgl_tunggakan_terakhir": null,
				"tgl_tutup_fasilitas_terakhir": null,
				"tunggakan_terburuk": 0
			},
			"fasilitas": [
				{
					"agunan": [
						{
							"alamat_agunan": "JALAN DUMMY 7705449706685359408",
							"bukti_kepemilikan": "NO DOKUMEN 7705449706685359408",
							"diasuransikan": "No",
							"id_jenis_agunan": "F2001",
							"id_jenis_pengikatan": "02",
							"id_lembaga_pemeringkat": "0",
							"id_status_agunan": "1",
							"keterangan": null,
							"kode_register_atau_nomor_agunan": "7705449706685359408",
							"nama_pemilik_agunan": "PEMILIK AGUNAN 7705449706685359408",
							"nama_penilai_independen": "PENILAI 7705449706685359408",
							"nilai_agunan_menurut_pelapor": 13836540,
							"nilai_agunan_menurut_penilai_independen": 0,
							"nilai_agunan_sesuai_njop": 11069232,
							"peringkat_agunan": null,
							"persentase_paripasu": null,
							"status_paripasu": "T",
							"tanggal_pengikatan": "2024-01-31",
							"tanggal_penilaian_agunan_menurut_pelapor": "2024-01-31",
							"tanggal_penilaian_agunan_menurut_penilai_independen": null
						}
					],
					"alamat_penjamin": null,
					"baki_debet": 10900000,
					"denda": 0,
					"disbursement_date": "",
					"fase_fasilitas": 1,
					"frekuensi_perpanjangan_fasilitas": 0,
					"frekuensi_restrukturisasi": 0,
					"frekuensi_tunggakan": 0,
					"id_cara_restrukturisasi": "0",
					"id_jenis_fasilitas": "F01",
					"id_jenis_identitas": null,
					"id_jenis_kredit": "F01-P08",
					"id_jenis_pelapor": -480784535,
					"id_jenis_penggunaan": "F01-3",
					"id_jenis_suku_bunga_atau_imbalan": "5",
					"id_kabupaten_kota": "3303",
					"id_kolektabilitas": "1",
					"id_kondisi": 1,
					"id_kredit_program_pemerintah": 1,
					"id_orientasi_penggunaan": "3",
					"id_pelapor": 584406412,
					"id_sebab_macet": "0",
					"id_sektor_ekonomi": "004190",
					"id_sifat_kredit": "9",
					"id_skim_pembiayaan": "100",
					"id_valuta": "IDR",
					"jml_hari_tunggakan_terburuk": 0,
					"jml_hari_tunggakan_terburuk_12bln": 0,
					"jml_tunggakan": 0,
					"jumlah_hari_tunggakan": 0,
					"keterangan": "",
					"keterangan_sebab_macet": "",
					"koletabilitas_terburuk": null,
					"koletabilitas_terburuk_12bln": null,
					"listing": "",
					"nama_lengkap_penjamin": null,
					"nama_penjamin_sesuai_identitas": null,
					"nilai_dalam_mata_uang_asal": 0,
					"nilai_pasar": 0,
					"nilai_perolehan": 0,
					"nilai_proyek": 0,
					"nominal_tunggakan": 0,
					"nomor_akad_akhir": "6233088051913989995",
					"nomor_akad_awal": "6233088051913989995",
					"nomor_identitas_penjamin": null,
					"nomor_rekening_fasilitas": "5847217818177690657",
					"penjamin": [],
					"peringkat_surat_berharga": "",
					"plafon": 10900000,
					"plafon_awal": 10900000,
					"realisasi_atau_pencairan_bulan_berjalan": 0,
					"riwayat_fasilitas": [
						{
							"baki_debet": 10900000,
							"denda": 0,
							"id_kolektabilitas": "1",
							"jumlah_hari_tunggakan": 0,
							"nominal_tunggakan": 0,
							"saldo_terutang": 10900000,
							"snapshot_order": 4,
							"status_tunggakan": "0",
							"suku_bunga_atau_imbalan": 19.33,
							"tahun_bulan_data": "2024-05-31",
							"tunggakan_bunga_atau_imbalan": 0,
							"tunggakan_pokok": "0.00"
						},
						{
							"baki_debet": 10900000,
							"denda": 0,
							"id_kolektabilitas": "1",
							"jumlah_hari_tunggakan": 0,
							"nominal_tunggakan": 0,
							"saldo_terutang": 10900000,
							"snapshot_order": 3,
							"status_tunggakan": "0",
							"suku_bunga_atau_imbalan": 19.33,
							"tahun_bulan_data": "2024-04-30",
							"tunggakan_bunga_atau_imbalan": 0,
							"tunggakan_pokok": "0.00"
						},
						{
							"baki_debet": 10900000,
							"denda": 0,
							"id_kolektabilitas": "1",
							"jumlah_hari_tunggakan": 0,
							"nominal_tunggakan": 0,
							"saldo_terutang": 10900000,
							"snapshot_order": 2,
							"status_tunggakan": "0",
							"suku_bunga_atau_imbalan": 19.33,
							"tahun_bulan_data": "2024-03-31",
							"tunggakan_bunga_atau_imbalan": 0,
							"tunggakan_pokok": "0.00"
						},
						{
							"baki_debet": 10900000,
							"denda": 0,
							"id_kolektabilitas": "1",
							"jumlah_hari_tunggakan": 0,
							"nominal_tunggakan": 0,
							"saldo_terutang": 10900000,
							"snapshot_order": 1,
							"status_tunggakan": "0",
							"suku_bunga_atau_imbalan": 19.33,
							"tahun_bulan_data": "2024-02-29",
							"tunggakan_bunga_atau_imbalan": 0,
							"tunggakan_pokok": "0.00"
						},
						{
							"baki_debet": 10900000,
							"denda": 0,
							"id_kolektabilitas": "1",
							"jumlah_hari_tunggakan": 0,
							"nominal_tunggakan": 0,
							"saldo_terutang": 10900000,
							"snapshot_order": 0,
							"status_tunggakan": "0",
							"suku_bunga_atau_imbalan": 19.33,
							"tahun_bulan_data": "2024-01-31",
							"tunggakan_bunga_atau_imbalan": 0,
							"tunggakan_pokok": "0.00"
						}
					],
					"saldo_terutang": 10900000,
					"suku_bunga_atau_imbalan": 19.33,
					"syndicated_loan": 0,
					"tahun_bulan_data": "2024-05-31",
					"tanggal_akad_akhir": "2024-01-31",
					"tanggal_akad_awal": "2024-01-31",
					"tanggal_akhir": "",
					"tanggal_awal_kredit_atau_pembiayaan": "2024-01-31",
					"tanggal_jatuh_tempo": "2024-05-25",
					"tanggal_keluar": "",
					"tanggal_kondisi": "",
					"tanggal_macet": "",
					"tanggal_mulai": "2024-01-31",
					"tanggal_pembelian": "",
					"tanggal_penerbitan": "",
					"tanggal_restrukturisasi_akhir": "",
					"tanggal_restrukturisasi_awal": "",
					"tanggal_wanprestasi": "",
					"tgl_tunggakan_terakhir": "",
					"tunggakan_bunga_atau_imbalan": 0,
					"tunggakan_pokok": 0,
					"tunggakan_terburuk": 0,
					"tunggakan_terburuk_12bln": 0
				}
			],
			"header": {
				"id_report": "548570",
				"tgl_permintaan": "2025-02-18T07:39:04.359236Z[GMT]",
				"username": "bfi_api"
			},
			"other_data": [],
			"pengurus": [],
			"permintaan_data": [
				{
					"deskripsi_jenis_ljk": "Member Mockrun Eagle",
					"id_jenis_pelapor": 117,
					"id_pelapor": "107",
					"id_tujuan_permintaan": 39,
					"nama_ljk": "BFI Finance Indonesia",
					"tgl_permintaan": "2025-02-18"
				},
				{
					"deskripsi_jenis_ljk": "Member Mockrun Eagle",
					"id_jenis_pelapor": 117,
					"id_pelapor": "107",
					"id_tujuan_permintaan": 39,
					"nama_ljk": "BFI Finance Indonesia",
					"tgl_permintaan": "2025-02-18"
				}
			],
			"riwayat_identitas_debitur": [
				{
					"id_elemen": 267913957,
					"id_element": null,
					"id_pelapor": 584406412,
					"tahun_bulan_data": "2024-05-31",
					"value": "Mom Chuumon Sepuluh Tiga"
				},
				{
					"id_elemen": -795716188,
					"id_element": null,
					"id_pelapor": 584406412,
					"tahun_bulan_data": "2024-05-31",
					"value": "1990-02-01"
				},
				{
					"id_elemen": 1266008634,
					"id_element": null,
					"id_pelapor": 584406412,
					"tahun_bulan_data": "2024-05-31",
					"value": "ChuumonSepuluhTiga@email.com"
				},
				{
					"id_elemen": -1727473890,
					"id_element": null,
					"id_pelapor": 584406412,
					"tahun_bulan_data": "2024-05-31",
					"value": "Chuumon Sepuluh Tiga"
				},
				{
					"id_elemen": -1680332394,
					"id_element": null,
					"id_pelapor": 584406412,
					"tahun_bulan_data": "2024-05-31",
					"value": "622153023593"
				},
				{
					"id_elemen": 1606855272,
					"id_element": null,
					"id_pelapor": 584406412,
					"tahun_bulan_data": "2024-05-31",
					"value": "6287821823493"
				},
				{
					"id_elemen": 910886720,
					"id_element": null,
					"id_pelapor": 584406412,
					"tahun_bulan_data": "2024-05-31",
					"value": "Jalan Dummy 0101010290023490"
				},
				{
					"id_elemen": -1465760571,
					"id_element": null,
					"id_pelapor": 584406412,
					"tahun_bulan_data": "2024-05-31",
					"value": "3175010102903493"
				},
				{
					"id_elemen": -1873769009,
					"id_element": null,
					"id_pelapor": 584406412,
					"tahun_bulan_data": "2024-05-31",
					"value": "3175010102903493"
				}
			],
			"summary_permintaan_data": [
				{
					"jml_pelapor_12bln": null,
					"jml_pelapor_1bln": null,
					"jml_pelapor_24bln": null,
					"jml_pelapor_3bln": null,
					"jml_pelapor_6bln": null,
					"jml_permintaan_12bln": null,
					"jml_permintaan_1bln": null,
					"jml_permintaan_24bln": null,
					"jml_permintaan_3bln": null,
					"jml_permintaan_6bln": null
				},
				{
					"jml_pelapor_12bln": null,
					"jml_pelapor_1bln": null,
					"jml_pelapor_24bln": null,
					"jml_pelapor_3bln": null,
					"jml_pelapor_6bln": null,
					"jml_permintaan_12bln": null,
					"jml_permintaan_1bln": null,
					"jml_permintaan_24bln": null,
					"jml_permintaan_3bln": null,
					"jml_permintaan_6bln": null
				},
				{
					"jml_pelapor_12bln": null,
					"jml_pelapor_1bln": null,
					"jml_pelapor_24bln": null,
					"jml_pelapor_3bln": null,
					"jml_pelapor_6bln": null,
					"jml_permintaan_12bln": null,
					"jml_permintaan_1bln": null,
					"jml_permintaan_24bln": null,
					"jml_permintaan_3bln": null,
					"jml_permintaan_6bln": null
				},
				{
					"jml_pelapor_12bln": null,
					"jml_pelapor_1bln": null,
					"jml_pelapor_24bln": null,
					"jml_pelapor_3bln": null,
					"jml_pelapor_6bln": null,
					"jml_permintaan_12bln": null,
					"jml_permintaan_1bln": null,
					"jml_permintaan_24bln": null,
					"jml_permintaan_3bln": null,
					"jml_permintaan_6bln": null
				},
				{
					"jml_pelapor_12bln": null,
					"jml_pelapor_1bln": null,
					"jml_pelapor_24bln": null,
					"jml_pelapor_3bln": null,
					"jml_pelapor_6bln": null,
					"jml_permintaan_12bln": null,
					"jml_permintaan_1bln": null,
					"jml_permintaan_24bln": null,
					"jml_permintaan_3bln": null,
					"jml_permintaan_6bln": null
				},
				{
					"jml_pelapor_12bln": null,
					"jml_pelapor_1bln": null,
					"jml_pelapor_24bln": null,
					"jml_pelapor_3bln": null,
					"jml_pelapor_6bln": null,
					"jml_permintaan_12bln": null,
					"jml_permintaan_1bln": null,
					"jml_permintaan_24bln": null,
					"jml_permintaan_3bln": null,
					"jml_permintaan_6bln": null
				},
				{
					"jml_pelapor_12bln": null,
					"jml_pelapor_1bln": null,
					"jml_pelapor_24bln": null,
					"jml_pelapor_3bln": null,
					"jml_pelapor_6bln": null,
					"jml_permintaan_12bln": null,
					"jml_permintaan_1bln": null,
					"jml_permintaan_24bln": null,
					"jml_permintaan_3bln": null,
					"jml_permintaan_6bln": null
				},
				{
					"jml_pelapor_12bln": null,
					"jml_pelapor_1bln": null,
					"jml_pelapor_24bln": null,
					"jml_pelapor_3bln": null,
					"jml_pelapor_6bln": null,
					"jml_permintaan_12bln": null,
					"jml_permintaan_1bln": null,
					"jml_permintaan_24bln": null,
					"jml_permintaan_3bln": null,
					"jml_permintaan_6bln": null
				},
				{
					"jml_pelapor_12bln": null,
					"jml_pelapor_1bln": null,
					"jml_pelapor_24bln": null,
					"jml_pelapor_3bln": null,
					"jml_pelapor_6bln": null,
					"jml_permintaan_12bln": null,
					"jml_permintaan_1bln": null,
					"jml_permintaan_24bln": null,
					"jml_permintaan_3bln": null,
					"jml_permintaan_6bln": null
				},
				{
					"jml_pelapor_12bln": null,
					"jml_pelapor_1bln": null,
					"jml_pelapor_24bln": null,
					"jml_pelapor_3bln": null,
					"jml_pelapor_6bln": null,
					"jml_permintaan_12bln": null,
					"jml_permintaan_1bln": null,
					"jml_permintaan_24bln": null,
					"jml_permintaan_3bln": null,
					"jml_permintaan_6bln": null
				},
				{
					"jml_pelapor_12bln": null,
					"jml_pelapor_1bln": null,
					"jml_pelapor_24bln": null,
					"jml_pelapor_3bln": null,
					"jml_pelapor_6bln": null,
					"jml_permintaan_12bln": null,
					"jml_permintaan_1bln": null,
					"jml_permintaan_24bln": null,
					"jml_permintaan_3bln": null,
					"jml_permintaan_6bln": null
				},
				{
					"jml_pelapor_12bln": null,
					"jml_pelapor_1bln": null,
					"jml_pelapor_24bln": null,
					"jml_pelapor_3bln": null,
					"jml_pelapor_6bln": null,
					"jml_permintaan_12bln": null,
					"jml_permintaan_1bln": null,
					"jml_permintaan_24bln": null,
					"jml_permintaan_3bln": null,
					"jml_permintaan_6bln": null
				}
			],
			"summary_riwayat_debitur": [
				{
					"jml_hari_tunggakan_terburuk": 100,
					"kolektabilitas_terburuk": "1",
					"tahun_bulan_data": "2024-05-31"
				},
				{
					"jml_hari_tunggakan_terburuk": 50,
					"kolektabilitas_terburuk": "1",
					"tahun_bulan_data": "2024-04-30"
				},
				{
					"jml_hari_tunggakan_terburuk": 1000,
					"kolektabilitas_terburuk": "1",
					"tahun_bulan_data": "2024-03-31"
				},
				{
					"jml_hari_tunggakan_terburuk": 500,
					"kolektabilitas_terburuk": "1",
					"tahun_bulan_data": "2024-02-29"
				},
				{
					"jml_hari_tunggakan_terburuk": 700,
					"kolektabilitas_terburuk": "1",
					"tahun_bulan_data": "2024-01-31"
				}
			]
		},
		"scoring": [
			{
				"id_pefindo": "` + pefindoID + `",
				"period": "2025-02-18",
				"pod": 4.33,
				"reason_code": [
					"NQS1"
				],
				"reason_desc": [
					"Dummy scoring"
				],
				"risk_grade": "A3",
				"risk_grade_desc": "Very Low Risk",
				"score": 664
			},
			{
				"id_pefindo": "` + pefindoID + `",
				"period": "2025-02-17",
				"pod": 4.33,
				"reason_code": [
					"NQS1"
				],
				"reason_desc": [
					"Dummy scoring"
				],
				"risk_grade": "A3",
				"risk_grade_desc": "Very Low Risk",
				"score": 664
			},
			{
				"id_pefindo": "` + pefindoID + `",
				"period": "2025-02-16",
				"pod": 4.33,
				"reason_code": [
					"NQS1"
				],
				"reason_desc": [
					"Dummy scoring"
				],
				"risk_grade": "A3",
				"risk_grade_desc": "Very Low Risk",
				"score": 664
			},
			{
				"id_pefindo": "` + pefindoID + `",
				"period": "2025-02-15",
				"pod": 4.33,
				"reason_code": [
					"NQS1"
				],
				"reason_desc": [
					"Dummy scoring"
				],
				"risk_grade": "A3",
				"risk_grade_desc": "Very Low Risk",
				"score": 664
			},
			{
				"id_pefindo": "` + pefindoID + `",
				"period": "2025-02-14",
				"pod": 4.33,
				"reason_code": [
					"NQS1"
				],
				"reason_desc": [
					"Dummy scoring"
				],
				"risk_grade": "A3",
				"risk_grade_desc": "Very Low Risk",
				"score": 664
			},
			{
				"id_pefindo": "` + pefindoID + `",
				"period": "2025-02-13",
				"pod": 4.33,
				"reason_code": [
					"NQS1"
				],
				"reason_desc": [
					"Dummy scoring"
				],
				"risk_grade": "A3",
				"risk_grade_desc": "Very Low Risk",
				"score": 664
			},
			{
				"id_pefindo": "` + pefindoID + `",
				"period": "2025-02-12",
				"pod": 4.33,
				"reason_code": [
					"NQS1"
				],
				"reason_desc": [
					"Dummy scoring"
				],
				"risk_grade": "A3",
				"risk_grade_desc": "Very Low Risk",
				"score": 664
			},
			{
				"id_pefindo": "` + pefindoID + `",
				"period": "2025-02-11",
				"pod": 4.33,
				"reason_code": [
					"NQS1"
				],
				"reason_desc": [
					"Dummy scoring"
				],
				"risk_grade": "A3",
				"risk_grade_desc": "Very Low Risk",
				"score": 664
			},
			{
				"id_pefindo": "` + pefindoID + `",
				"period": "2025-02-10",
				"pod": 5.76,
				"reason_code": [
					"NQS1"
				],
				"reason_desc": [
					"Dummy scoring"
				],
				"risk_grade": "A3",
				"risk_grade_desc": "Very Low Risk",
				"score": 651
			},
			{
				"id_pefindo": "` + pefindoID + `",
				"period": "2025-02-09",
				"pod": 5.76,
				"reason_code": [
					"NQS1"
				],
				"reason_desc": [
					"Dummy scoring"
				],
				"risk_grade": "A3",
				"risk_grade_desc": "Very Low Risk",
				"score": 651
			},
			{
				"id_pefindo": "` + pefindoID + `",
				"period": "2025-02-08",
				"pod": 5.76,
				"reason_code": [
					"NQS1"
				],
				"reason_desc": [
					"Dummy scoring"
				],
				"risk_grade": "A3",
				"risk_grade_desc": "Very Low Risk",
				"score": 651
			},
			{
				"id_pefindo": "` + pefindoID + `",
				"period": "2025-02-07",
				"pod": 5.76,
				"reason_code": [
					"NQS1"
				],
				"reason_desc": [
					"Dummy scoring"
				],
				"risk_grade": "A3",
				"risk_grade_desc": "Very Low Risk",
				"score": 651
			},
			{
				"id_pefindo": "` + pefindoID + `",
				"period": "2025-02-06",
				"pod": 5.76,
				"reason_code": [
					"NQS1"
				],
				"reason_desc": [
					"Dummy scoring"
				],
				"risk_grade": "A3",
				"risk_grade_desc": "Very Low Risk",
				"score": 651
			}
		]
	}`
	return jsonResp
}

func getRoundRobinPefindoIDs() string {
	// read file perindo_ids.txt
	// split by ,
	f, err := os.ReadFile("pefindo_ids.txt")
	if err != nil {
		log.Fatal(err)
	}
	pefindoIDs := strings.Split(string(f), ",")
	if len(pefindoIDs) == 0 {
		return ""
	}

	// pefindoIDs := []string{"102160861000014", "102160861000015"}
	idx := 0
	if counter%len(pefindoIDs) != 0 {
		idx = 1
	}
	counter++

	return pefindoIDs[idx]
}
