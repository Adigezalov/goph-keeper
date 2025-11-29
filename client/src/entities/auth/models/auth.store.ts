import { makeAutoObservable, runInAction } from 'mobx'

import { db } from '@shared/db'
import { TStoreLogic } from '@shared/store'
import { getAccessToken, removeTokens } from '@shared/tokens'

import {
	loginApi,
	logoutAllApi,
	logoutApi,
	refreshApi,
	registrationApi,
	resendCodeApi,
	verifyEmailApi,
} from '../api'
import { TAuth, TResendCode, TVerifyEmail } from '../types'

export class AuthStore {
	auth = false
	awaitingEmailVerification = false
	verificationEmail = ''

	isLoginLoading = false
	isRegistrationLoading = false
	isVerifyEmailLoading = false
	isResendCodeLoading = false
	isCheckLoading = false
	isLogoutLoading = false

	rootStore: TStoreLogic

	constructor(rootStore: TStoreLogic) {
		this.rootStore = rootStore
		makeAutoObservable(this, {}, { autoBind: true })
	}

	login = async (data: TAuth) => {
		try {
			this.isLoginLoading = true

			await loginApi(data)

			runInAction(() => {
				this.auth = true
			})
		} catch (_) {
			runInAction(() => {
				this.auth = false
			})
		} finally {
			runInAction(() => {
				this.isLoginLoading = false
			})
		}
	}

	registration = async (data: TAuth) => {
		try {
			this.isRegistrationLoading = true

			await registrationApi(data)

			runInAction(() => {
				this.awaitingEmailVerification = true
				this.verificationEmail = data.email || ''
				this.auth = false
			})
		} catch (_) {
			runInAction(() => {
				this.auth = false
				this.awaitingEmailVerification = false
			})
		} finally {
			runInAction(() => {
				this.isRegistrationLoading = false
			})
		}
	}

	verifyEmail = async (data: TVerifyEmail) => {
		try {
			this.isVerifyEmailLoading = true

			await verifyEmailApi(data)

			runInAction(() => {
				this.auth = true
				this.awaitingEmailVerification = false
				this.verificationEmail = ''
			})
		} catch (_) {
			runInAction(() => {
				this.auth = false
			})
			throw _
		} finally {
			runInAction(() => {
				this.isVerifyEmailLoading = false
			})
		}
	}

	resendCode = async (data: TResendCode) => {
		try {
			this.isResendCodeLoading = true

			await resendCodeApi(data)
		} catch (_) {
			throw _
		} finally {
			runInAction(() => {
				this.isResendCodeLoading = false
			})
		}
	}

	refresh = async () => {
		try {
			if (this.isCheckLoading) return

			this.isCheckLoading = true

			const accessToken = getAccessToken()

			if (!accessToken) {
				runInAction(() => {
					this.auth = false
				})

				await this.logout()

				return
			}

			await refreshApi()

			runInAction(() => {
				this.auth = true
			})
		} catch (_) {
			runInAction(() => {
				this.auth = false
			})

			await this.logout()
		} finally {
			runInAction(() => {
				this.isCheckLoading = false
			})
		}
	}

	logout = async () => {
		try {
			this.isLogoutLoading = true

			const accessToken = getAccessToken()

			if (accessToken) await logoutApi()

			removeTokens()

			runInAction(() => {
				this.clearAuthStore()
			})
		} finally {
			runInAction(() => {
				this.isLogoutLoading = false
			})
		}
	}

	logoutAll = async () => {
		try {
			this.isLogoutLoading = true

			const accessToken = getAccessToken()

			if (accessToken) await logoutAllApi()

			removeTokens()

			runInAction(() => {
				this.clearAuthStore()
			})
		} finally {
			runInAction(() => {
				this.isLogoutLoading = false
			})
		}
	}

	initAuthStore = (): void => {
		void this.refresh()
	}

	clearAuthStore = () => {
		this.auth = false
		this.awaitingEmailVerification = false
		this.verificationEmail = ''

		this.isLoginLoading = false
		this.isRegistrationLoading = false
		this.isVerifyEmailLoading = false
		this.isResendCodeLoading = false
		this.isCheckLoading = false
		this.isLogoutLoading = false

		// Очистка IndexedDB
		void db.clearAllData()

		this.rootStore.cryptoKey.clearCryptoStore()
		this.rootStore.networkStatus.clearNetworkStatusStore()
		this.rootStore.realtime.clearRealtimeStore()
		this.rootStore.serverStatus.clearServerStatusStore()
		this.rootStore.secretsPage.clearSecretsPageStore()
	}
}
