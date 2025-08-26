package database

func RunMigrations() {
	EnablePgCrypto(PG_Client)
	CreateUserTable(PG_Client)
}
