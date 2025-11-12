import i18next from 'i18next'
import { makeAutoObservable, runInAction } from 'mobx'

import { TStoreLogic } from '@shared/store'
import { showToastNotification } from '@shared/toast-notification'
import { TOAST_SEVERITY } from '@shared/uikit/toast'

import {
	exportCryptoKey,
	getCryptoKey,
	importCryptoKey,
	removeCryptoKey,
	setCryptoKey,
} from '../utils'

export class CryptoStore {
	cryptoKey?: CryptoKey = undefined

	visibleCryptoModal = false

	rootStore: TStoreLogic

	constructor(rootStore: TStoreLogic) {
		this.rootStore = rootStore
		makeAutoObservable(this, {}, { autoBind: true })
	}

	setVisibleCryptoModal = (visible: boolean) => {
		this.visibleCryptoModal = visible
	}

	setCryptoKey = async (keyString?: string) => {
		try {
			if (!keyString) return

			const key = await importCryptoKey(keyString)

			setCryptoKey(keyString)

			runInAction(() => {
				this.cryptoKey = key

				this.visibleCryptoModal = false
			})

			return true
		} catch (e) {
			console.log('Ошибка установки ключа', e)

			return false
		}
	}

	getCryptoKeyString = async (): Promise<string | null> => {
		try {
			if (!this.cryptoKey) {
				return null
			}

			return await exportCryptoKey(this.cryptoKey)
		} catch (e) {
			console.log('Ошибка экспорта ключа', e)

			return null
		}
	}

	copyCryptoKey = async (): Promise<boolean> => {
		try {
			const keyString = await this.getCryptoKeyString()

			if (!keyString) {
				console.error('Ошибка при копировании: ключ не найден')
				return false
			}

			await navigator.clipboard.writeText(keyString)

			showToastNotification({
				message: i18next.t('copy_crypto_key_success'),
				header: i18next.t('info'),
				severity: TOAST_SEVERITY.INFO,
			})

			return true
		} catch (err) {
			console.error('Ошибка при копировании: ', err)
			return false
		}
	}

	initCryptoStore = async () => {
		try {
			const keyString = getCryptoKey()

			if (!keyString) this.visibleCryptoModal = true

			if (keyString) {
				const key = await importCryptoKey(keyString)
				runInAction(() => {
					this.cryptoKey = key
				})
			}
		} catch (e) {
			console.log('Ошибка загрузки ключа из localStorage', e)

			removeCryptoKey()

			runInAction(() => {
				this.cryptoKey = undefined
			})
		}
	}

	clearCryptoStore = () => {
		this.cryptoKey = undefined
		this.visibleCryptoModal = false
	}
}
