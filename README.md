# Example CRUD - CouchDB

## Program ini merupakan contoh implementasi CRUD menggunakan CouchDB, dengan fitur utama sinkronisasi otomatis antar node.

## Fitur Utama:
- Sinkronisasi Otomatis Antar Node: Setiap perubahan pada database lokal di satu node akan secara otomatis tersinkronisasi ke database lokal di node lain.

## Contoh Skenario:
- NODE1 melakukan operasi Update pada database lokalnya.
- Secara otomatis, database lokal di NODE2 dan node lainnya juga akan menerima perubahan tersebut.

## Cara Kerja Singkat:
Sinkronisasi ini memanfaatkan fitur replication dari CouchDB.

## Persyaratan:
- CouchDB harus diaktifkan pada semua node yang ingin tersinkronisasi.
- Konfigurasi replicator pada CouchDB harus disesuaikan untuk mendukung sinkronisasi otomatis.

source: https://github.com/fjl/go-couchdb
      : https://github.com/apache/couchdb
      : https://docs.couchdb.org/en/stable/
