package main

type Config struct {
	token    string
	api_keys []string
	admins   []int
	VIPs     []int
}

var config = Config{
	token:    "6071720533:AAGBSSZs3t8wIJ7oIKQEnHpeS34R8Xo684U", //"6285536369:AAEkcIv6o2KVRfhl_r8mZkrU45kHh1h76tw",
	api_keys: []string{"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", "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", "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", "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", "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"},
	admins:   []int{5429321679},
	VIPs:     []int{938320319, 1073960679, 1329919004},
}
