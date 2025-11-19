import { makeAutoObservable, runInAction } from 'mobx'

import { TStoreLogic } from '@shared/store'
import { getAccessToken, removeTokens } from '@shared/tokens'

import { loginApi, logoutAllApi, logoutApi, refreshApi, registrationApi } from '../api'
import { TAuth } from '../types'

export class AuthStore {
	auth = false

	isLoginLoading = false
	isRegistrationLoading = false
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
				this.auth = true
			})
		} catch (_) {
			runInAction(() => {
				this.auth = false
			})
		} finally {
			runInAction(() => {
				this.isRegistrationLoading = false
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

		this.isLoginLoading = false
		this.isRegistrationLoading = false
		this.isCheckLoading = false
		this.isLogoutLoading = false

		this.rootStore.cryptoKey.clearCryptoStore()
		this.rootStore.networkStatus.clearNetworkStatusStore()
		this.rootStore.realtime.clearRealtimeStore()
		this.rootStore.serverStatus.clearServerStatusStore()
		this.rootStore.secretsPage.clearSecretsPageStore()
	}
}
