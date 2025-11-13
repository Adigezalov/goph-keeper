import { makeAutoObservable, runInAction } from 'mobx'
import { v4 as uuidv4 } from 'uuid'

import { db } from '@shared/db'
import { TStoreLogic } from '@shared/store'
import { showToastNotification } from '@shared/toast-notification'
import { TOAST_SEVERITY } from '@shared/uikit/toast'

import { TSecret, TSecretForSave } from '../types'

export class SecretsPageStore {
	secrets: TSecret[] = []

	syncStatus: 'idle' | 'syncing' | 'error' = 'idle'
	lastSyncTime: Date | null = null
	unsyncedCount = 0

	isLoading = false
	isCreating = false
	isUpdating = false
	isDeleting = false

	rootStore: TStoreLogic

	constructor(rootStore: TStoreLogic) {
		this.rootStore = rootStore
		makeAutoObservable(this, {}, { autoBind: true })
	}

	async createSecret(data: { secret: TSecretForSave; cb: () => void }) {
		try {
			this.isCreating = true

			const { login, password, metadata, binaryData } = data.secret

			// Шифруем все поля кроме metadata
			const encryptedLogin = await this.encryptData(login.trim())
			const encryptedPassword = await this.encryptData(password)
			const encryptedBinaryData = binaryData
				? await this.encryptBinaryData(binaryData)
				: undefined

			// Создаем локально
			const newSecret: TSecret = {
				localId: uuidv4(),
				login: encryptedLogin,
				password: encryptedPassword,
				metadata: metadata || {},
				binaryData: encryptedBinaryData,
				version: 1,
				syncStatus: 'pending',
				createdAt: Date.now(),
				updatedAt: Date.now(),
			}

			await db.secrets.add(newSecret)

			// Добавляем в очередь синхронизации
			// await syncService.addToQueue('create', newSecret)

			// Обновляем UI
			runInAction(() => {
				this.secrets = [newSecret, ...this.secrets]
			})

			data.cb()

			// Синхронизируем если онлайн
			// if (navigator.onLine) {
			// 	await this.sync()
			// }

			// await this.updateUnsyncedCount()

			return newSecret
		} catch (error) {
			console.error('Ошибка создания секрета:', error)
			showToastNotification({
				message: error instanceof Error ? error.message : 'Неизвестная ошибка',
				header: 'Ошибка создания секрета',
				severity: TOAST_SEVERITY.ERROR,
			})
		} finally {
			runInAction(() => {
				this.isCreating = false
			})
		}
	}

	async updateSecret(data: { localId: string; secret: TSecretForSave; cb: () => void }) {
		try {
			this.isUpdating = true

			const { localId, secret, cb } = data
			const { login, password, metadata, binaryData } = secret

			const existingSecret = await db.secrets.get(localId)
			if (!existingSecret) return

			// Шифруем все поля кроме metadata
			const encryptedLogin = await this.encryptData(login.trim())
			const encryptedPassword = await this.encryptData(password)
			const encryptedBinaryData = binaryData
				? await this.encryptBinaryData(binaryData)
				: undefined

			const updatedSecret: TSecret = {
				...existingSecret,
				login: encryptedLogin,
				password: encryptedPassword,
				metadata: metadata || {},
				binaryData: encryptedBinaryData,
				version: existingSecret.version + 1,
				syncStatus: 'pending',
				updatedAt: Date.now(),
			}

			await db.secrets.put(updatedSecret)

			// await syncService.addToQueue('update', updatedSecret)

			runInAction(() => {
				this.secrets = this.secrets.map((secret) => {
					if (secret.localId === localId) {
						return updatedSecret
					}
					return secret
				})

				cb()
			})

			// if (navigator.onLine) {
			// 	await this.sync()
			// }

			// await this.updateUnsyncedCount()
		} catch (error) {
			console.error('Ошибка обновления секрета:', error)
			showToastNotification({
				message: error instanceof Error ? error.message : 'Неизвестная ошибка',
				header: 'Ошибка обновления секрета',
				severity: TOAST_SEVERITY.ERROR,
			})
		} finally {
			runInAction(() => {
				this.isUpdating = false
			})
		}
	}

	async deleteSecret(id: string) {
		try {
			this.isDeleting = true

			const existingSecret = await db.secrets.get(id)
			if (!existingSecret) return

			// Если секрет еще не синхронизирован с сервером, удаляем полностью
			if (!existingSecret.id) {
				await db.secrets.delete(id)

				// Удаляем из очереди синхронизации
				await db.syncQueue.where('secretId').equals(id).delete()

				// Обновляем UI
				runInAction(() => {
					this.secrets = this.secrets.filter((s) => s.localId !== id)
				})
			} else {
				// Если секрет синхронизирован, делаем soft delete
				const deletedSecret: TSecret = {
					...existingSecret,
					syncStatus: 'deleted',
					deletedAt: Date.now(),
					updatedAt: Date.now(),
				}

				await db.secrets.put(deletedSecret)

				// Добавляем в очередь синхронизации
				// await syncService.addToQueue('delete', deletedSecret)

				// Обновляем UI (скрываем удаленный секрет)
				runInAction(() => {
					this.secrets = this.secrets.filter((s) => s.localId !== id)
				})

				// Синхронизируем если онлайн
				// if (navigator.onLine) {
				// 	await this.sync()
				// }
			}

			// await this.updateUnsyncedCount()
		} catch (error) {
			console.error('Ошибка удаления секрета:', error)
			showToastNotification({
				message: error instanceof Error ? error.message : 'Неизвестная ошибка',
				header: 'Ошибка удаления секрета',
				severity: TOAST_SEVERITY.ERROR,
			})
		} finally {
			runInAction(() => {
				this.isDeleting = false
			})
		}
	}

	async loadSecrets() {
		try {
			this.isLoading = true

			const secrets = await db.secrets.filter((secret) => !secret.deletedAt).toArray()

			runInAction(() => {
				this.secrets = secrets
			})
		} catch (error) {
			console.error('Ошибка загрузки секретов:', error)
		} finally {
			runInAction(() => {
				this.isLoading = false
			})
		}
	}

	// Полная синхронизация
	// async sync() {
	// 	try {
	// 		runInAction(() => {
	// 			this.syncStatus = 'syncing'
	// 		})
	//
	// 		await syncService.sync()
	//
	// 		// Перезагружаем секреты после синхронизации
	// 		await this.loadSecrets()
	//
	// 		runInAction(() => {
	// 			this.syncStatus = 'idle'
	// 			this.lastSyncTime = new Date()
	// 		})
	//
	// 		await this.updateUnsyncedCount()
	// 	} catch (error) {
	// 		console.error('Ошибка синхронизации:', error)
	// 		runInAction(() => {
	// 			this.syncStatus = 'error'
	// 		})
	//
	// 	}
	// }

	// Обновить счетчик несинхронизированных изменений
	// async updateUnsyncedCount() {
	// 	const count = await syncService.getUnsyncedCount()
	// 	runInAction(() => {
	// 		this.unsyncedCount = count
	// 	})
	// }

	private async encryptData(data: string): Promise<string> {
		const cryptoKey = this.rootStore.cryptoKey.cryptoKey

		if (!cryptoKey) {
			throw new Error('Ключ шифрования не установлен')
		}

		// Генерируем IV (initialization vector)
		const iv = window.crypto.getRandomValues(new Uint8Array(12))

		// Шифруем
		const encoder = new TextEncoder()
		const encodedData = encoder.encode(data)

		const encrypted = await window.crypto.subtle.encrypt(
			{
				name: 'AES-GCM',
				iv,
			},
			cryptoKey,
			encodedData,
		)

		// Объединяем IV и зашифрованные данные
		const combined = new Uint8Array(iv.length + encrypted.byteLength)
		combined.set(iv, 0)
		combined.set(new Uint8Array(encrypted), iv.length)

		// Конвертируем в base64
		return btoa(String.fromCharCode(...combined))
	}

	async decryptData(encryptedData: string): Promise<string> {
		const cryptoKey = this.rootStore.cryptoKey.cryptoKey

		if (!cryptoKey) {
			throw new Error('Ключ шифрования не установлен')
		}

		// Декодируем из base64
		const combined = Uint8Array.from(atob(encryptedData), (c) => c.charCodeAt(0))

		// Извлекаем IV и зашифрованные данные
		const iv = combined.slice(0, 12)
		const encrypted = combined.slice(12)

		// Расшифровываем
		const decrypted = await window.crypto.subtle.decrypt(
			{
				name: 'AES-GCM',
				iv,
			},
			cryptoKey,
			encrypted,
		)

		// Конвертируем в строку
		const decoder = new TextDecoder()
		return decoder.decode(decrypted)
	}

	private async encryptBinaryData(data: Uint8Array): Promise<Uint8Array> {
		const cryptoKey = this.rootStore.cryptoKey.cryptoKey

		if (!cryptoKey) {
			throw new Error('Ключ шифрования не установлен')
		}

		// Генерируем IV
		const iv = window.crypto.getRandomValues(new Uint8Array(12))

		// Шифруем
		const encrypted = await window.crypto.subtle.encrypt(
			{
				name: 'AES-GCM',
				iv,
			},
			cryptoKey,
			data.buffer as ArrayBuffer,
		)

		// Объединяем IV и зашифрованные данные
		const combined = new Uint8Array(iv.length + encrypted.byteLength)
		combined.set(iv, 0)
		combined.set(new Uint8Array(encrypted), iv.length)

		return combined
	}

	async decryptBinaryData(encryptedData: Uint8Array): Promise<Uint8Array> {
		const cryptoKey = this.rootStore.cryptoKey.cryptoKey

		if (!cryptoKey) {
			throw new Error('Ключ шифрования не установлен')
		}

		// Извлекаем IV и зашифрованные данные
		const iv = encryptedData.slice(0, 12)
		const encrypted = encryptedData.slice(12)

		// Расшифровываем
		const decrypted = await window.crypto.subtle.decrypt(
			{
				name: 'AES-GCM',
				iv,
			},
			cryptoKey,
			encrypted.buffer as ArrayBuffer,
		)

		return new Uint8Array(decrypted)
	}

	async initStore() {
		await this.loadSecrets()
		// await this.updateUnsyncedCount()

		// Попытаться синхронизировать при инициализации
		// if (navigator.onLine) {
		// 	await this.sync().catch((error) => {
		// 		console.log('Не удалось синхронизировать при инициализации:', error)
		// 	})
		// }
	}

	clearStore() {
		this.secrets = []
		this.syncStatus = 'idle'
		this.lastSyncTime = null
		this.unsyncedCount = 0
		this.isLoading = false
		this.isCreating = false
		this.isUpdating = false
		this.isDeleting = false
	}
}
