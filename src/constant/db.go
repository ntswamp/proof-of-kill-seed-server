package constant

const DbIdolverse string = "idolverse"

func GetDbConnectionSettings() map[string]map[string]string {
	switch SERVER_TYPE {
	case "development":
		return map[string]map[string]string{
			DbIdolverse: {
				"Host":     "",
				"Port":     "5432",
				"Database": "idolverse",
				"User":     "postgres",
				"Password": "password",
			},
		}
	case "production", "management":
		return map[string]map[string]string{
			DbIdolverse: {
				"Host":     "",
				"Port":     "",
				"Database": "idolverse",
				"User":     "postgres",
				"Password": "password",
			},
		}
	//local
	default:
		return map[string]map[string]string{
			DbIdolverse: {
				"Host":     "db",
				"Port":     "5432",
				"Database": "idolverse",
				"User":     "admin",
				"Password": "password",
			},
		}
	}
}
