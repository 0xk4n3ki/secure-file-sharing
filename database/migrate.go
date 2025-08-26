package database

func RunMigrations() {
	EnablePgCrypto(PG_Client)
	CreateUserTable(PG_Client)
	CreateFileTable(PG_Client)
	CreateFileAccessTable(PG_Client)
}
