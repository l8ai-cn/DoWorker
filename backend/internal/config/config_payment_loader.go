package config

func loadPaymentConfig() PaymentConfig {
	return PaymentConfig{
		DeploymentType: DeploymentType(getEnv("DEPLOYMENT_TYPE", "global")),
		MockEnabled:    getEnvBool("PAYMENT_MOCK", false),
		MockBaseURL:    getEnv("PAYMENT_MOCK_BASE_URL", ""),
		Stripe: StripeConfig{
			SecretKey:      getEnv("STRIPE_SECRET_KEY", ""),
			PublishableKey: getEnv("STRIPE_PUBLISHABLE_KEY", ""),
			WebhookSecret:  getEnv("STRIPE_WEBHOOK_SECRET", ""),
		},
		LemonSqueezy: LemonSqueezyConfig{
			APIKey:        getEnv("LEMONSQUEEZY_API_KEY", ""),
			StoreID:       getEnv("LEMONSQUEEZY_STORE_ID", ""),
			WebhookSecret: getEnv("LEMONSQUEEZY_WEBHOOK_SECRET", ""),
		},
		Alipay: AlipayConfig{
			AppID:           getEnv("ALIPAY_APP_ID", ""),
			PrivateKey:      getEnv("ALIPAY_PRIVATE_KEY", ""),
			AlipayPublicKey: getEnv("ALIPAY_PUBLIC_KEY", ""),
			IsSandbox:       getEnvBool("ALIPAY_SANDBOX", false),
		},
		WeChat: WeChatConfig{
			AppID:     getEnv("WECHAT_APP_ID", ""),
			MchID:     getEnv("WECHAT_MCH_ID", ""),
			APIKey:    getEnv("WECHAT_API_KEY", ""),
			APIv3Key:  getEnv("WECHAT_APIV3_KEY", ""),
			CertPath:  getEnv("WECHAT_CERT_PATH", ""),
			KeyPath:   getEnv("WECHAT_KEY_PATH", ""),
			IsSandbox: getEnvBool("WECHAT_SANDBOX", false),
		},
		License: LicenseConfig{
			PublicKeyPath:    getEnv("LICENSE_PUBLIC_KEY_PATH", ""),
			LicenseFilePath:  getEnv("LICENSE_FILE_PATH", ""),
			LicenseServerURL: getEnv("LICENSE_SERVER_URL", ""),
		},
	}
}
