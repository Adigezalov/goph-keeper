import { makeAutoObservable, reaction, runInAction } from 'mobx'
import { v4 as uuidv4 } from 'uuid'

import { db } from '@shared/db'
import { TStoreLogic } from '@shared/store'
import { showToastNotification } from '@shared/toast-notification'
import { TOAST_SEVERITY } from '@shared/uikit/toast'

import {
	createSecretApi,
	deleteSecretApi,
	syncSecretsApi,
	updateSecretApi,
} from '../api'
import {
	TCreateSecretRequest,
	TSecret,
	TSecretForSave,
	TSecretResponse,
	TUpdateSecretRequest,
} from '../types'

export class SecretsPageStore {
	secrets: TSecret[] = []

	syncStatus: 'idle' | 'syncing' | 'error' = 'idle'
	lastSyncTime: string | null = null // ISO timestamp для запроса синхронизации
	lastSyncDate: Date | null = null // Дата последней синхронизации для UI
	unsyncedCount = 0

	isLoading = false
	isCreating = false
	isUpdating = false
	isDeleting = false

	rootStore: TStoreLogic
	private disposeReactions: (() => void)[] = []

	constructor(rootStore: TStoreLogic) {
		this.rootStore = rootStore
		makeAutoObservable(this, {}, { autoBind: true })
		this.setupSyncReactions()
	}

	// Настройка реакций на изменения статуса сети и сервера
	private setupSyncReactions() {
		// Реакция на изменение статуса сети
		const networkReaction = reaction(
			() => this.rootStore.networkStatus.isOnline,
			async (isOnline) => {
				if (isOnline) {
					// При восстановлении сети немедленно проверяем статус сервера
					await this.rootStore.serverStatus.checkStatus()
					
					// Если есть несинхронизированные данные и сервер доступен, синхронизируем
					if (this.unsyncedCount > 0 && this.rootStore.serverStatus.status) {
						void this.sync()
					}
				}
			},
		)

		// Реакция на изменение статуса сервера
		const serverReaction = reaction(
			() => this.rootStore.serverStatus.status,
			(status) => {
				if (status && this.rootStore.networkStatus.isOnline && this.unsyncedCount > 0) {
					void this.sync()
				}
			},
		)

		// Реакция на появление несинхронизированных секретов
		// Если есть несинхронизированные секреты и доступ к сети/серверу, запускаем синхронизацию
		const unsyncedReaction = reaction(
			() => this.unsyncedCount,
			(count) => {
				if (count > 0 && this.canSync()) {
					void this.sync()
				}
			},
		)

		this.disposeReactions.push(networkReaction, serverReaction, unsyncedReaction)
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

			// Обновляем UI
			runInAction(() => {
				this.secrets = [newSecret, ...this.secrets]
			})

			data.cb()

			// Синхронизируем если онлайн
			if (this.canSync()) {
				void this.sync()
			}

			await this.updateUnsyncedCount()

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
				// Версию не меняем - её обновит сервер при синхронизации
				syncStatus: 'pending',
				updatedAt: Date.now(),
			}

			await db.secrets.put(updatedSecret)

			runInAction(() => {
				this.secrets = this.secrets.map((secret) => {
					if (secret.localId === localId) {
						return updatedSecret
					}
					return secret
				})

				cb()
			})

			// Синхронизируем если онлайн
			if (this.canSync()) {
				void this.sync()
			}

			await this.updateUnsyncedCount()
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

				// Обновляем UI (скрываем удаленный секрет)
				runInAction(() => {
					this.secrets = this.secrets.filter((s) => s.localId !== id)
				})

				// Синхронизируем если онлайн
				if (this.canSync()) {
					void this.sync()
				}
			}

			await this.updateUnsyncedCount()
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

			// Загружаем только неудаленные секреты
			const secrets = await db.secrets
				.filter((secret) => !secret.deletedAt && secret.syncStatus !== 'deleted')
				.toArray()

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

	// Проверка возможности синхронизации
	private canSync(): boolean {
		return (
			this.rootStore.networkStatus.isOnline &&
			this.rootStore.serverStatus.status &&
			this.syncStatus !== 'syncing'
		)
	}

	// Полная синхронизация
	async sync() {
		if (!this.canSync()) {
			return
		}

		try {
			runInAction(() => {
				this.syncStatus = 'syncing'
			})

			// 1. Отправляем локальные изменения на сервер
			await this.pushLocalChanges()

			// 2. Получаем изменения с сервера
			await this.pullServerChanges()

			// 3. Перезагружаем секреты после синхронизации
			await this.loadSecrets()

			runInAction(() => {
				this.syncStatus = 'idle'
				this.lastSyncDate = new Date()
			})

			await this.updateUnsyncedCount()
		} catch (error) {
			console.error('Ошибка синхронизации:', error)
			runInAction(() => {
				this.syncStatus = 'error'
			})
		}
	}

	// Отправка локальных изменений на сервер
	private async pushLocalChanges() {
		// Получаем все несинхронизированные секреты
		const pendingSecrets = await db.secrets
			.filter((secret) => secret.syncStatus === 'pending' || secret.syncStatus === 'deleted')
			.toArray()

		for (const secret of pendingSecrets) {
			try {
				if (secret.syncStatus === 'deleted' && secret.id) {
					// Удаляем на сервере
					await deleteSecretApi(secret.id)
					// Удаляем из локальной базы
					await db.secrets.delete(secret.localId)
				} else if (!secret.id) {
					// Создаем новый секрет на сервере
					const response = await createSecretApi(this.secretToCreateRequest(secret))
					// Обновляем локальный секрет с id и version с сервера
					await db.secrets.put({
						...secret,
						id: response.data.id,
						version: response.data.version,
						syncStatus: 'synced',
					})
				} else {
					// Обновляем существующий секрет на сервере
					const response = await updateSecretApi(
						secret.id,
						this.secretToUpdateRequest(secret),
					)
					// Обновляем локальный секрет с новой version с сервера
					await db.secrets.put({
						...secret,
						version: response.data.version,
						syncStatus: 'synced',
					})
				}
			} catch (error) {
				console.error('Ошибка отправки секрета на сервер:', error)
				// Продолжаем синхронизацию других секретов
			}
		}
	}

	// Получение изменений с сервера
	private async pullServerChanges() {
		try {
			const response = await syncSecretsApi(this.lastSyncTime || undefined)

			// Обрабатываем каждый секрет с сервера
			for (const serverSecret of response.data.secrets) {
				await this.applyServerSecret(serverSecret)
			}

			// Сохраняем время синхронизации для следующего запроса
			runInAction(() => {
				this.lastSyncTime = response.data.server_time
			})
		} catch (error) {
			console.error('Ошибка получения данных с сервера:', error)
			throw error
		}
	}

	// Применение секрета с сервера к локальной базе
	private async applyServerSecret(serverSecret: TSecretResponse) {
		// Ищем локальный секрет по server id
		const existingSecret = await db.secrets
			.filter((s) => s.id === serverSecret.id)
			.first()

		// Конвертируем binary_data из base64 в Uint8Array если есть
		const binaryData = serverSecret.binary_data
			? this.base64ToUint8Array(serverSecret.binary_data)
			: undefined

		if (serverSecret.deleted_at) {
			// Секрет удален на сервере
			if (existingSecret) {
				await db.secrets.delete(existingSecret.localId)
			}
		} else if (existingSecret) {
			// Обновляем существующий секрет (без конфликтов, просто берем версию с сервера)
			await db.secrets.put({
				...existingSecret,
				login: serverSecret.login,
				password: serverSecret.password,
				metadata: serverSecret.metadata,
				binaryData,
				version: serverSecret.version,
				syncStatus: 'synced',
				updatedAt: new Date(serverSecret.updated_at).getTime(),
			})
		} else {
			// Создаем новый локальный секрет
			const newSecret: TSecret = {
				localId: uuidv4(),
				id: serverSecret.id,
				login: serverSecret.login,
				password: serverSecret.password,
				metadata: serverSecret.metadata,
				binaryData,
				version: serverSecret.version,
				syncStatus: 'synced',
				createdAt: new Date(serverSecret.created_at).getTime(),
				updatedAt: new Date(serverSecret.updated_at).getTime(),
			}
			await db.secrets.add(newSecret)
		}
	}

	// Обновить счетчик несинхронизированных изменений
	async updateUnsyncedCount() {
		const count = await db.secrets
			.filter((secret) => secret.syncStatus === 'pending' || secret.syncStatus === 'deleted')
			.count()
		runInAction(() => {
			this.unsyncedCount = count
		})
	}

	// Конвертация локального секрета в формат для создания на сервере
	private secretToCreateRequest(secret: TSecret): TCreateSecretRequest {
		return {
			login: secret.login,
			password: secret.password,
			metadata: secret.metadata,
			binary_data: secret.binaryData ? this.uint8ArrayToBase64(secret.binaryData) : undefined,
		}
	}

	// Конвертация локального секрета в формат для обновления на сервере
	private secretToUpdateRequest(secret: TSecret): TUpdateSecretRequest {
		return {
			login: secret.login,
			password: secret.password,
			metadata: secret.metadata,
			binary_data: secret.binaryData ? this.uint8ArrayToBase64(secret.binaryData) : undefined,
			version: secret.version,
		}
	}

	// Конвертация Uint8Array в base64
	private uint8ArrayToBase64(data: Uint8Array): string {
		// Для больших массивов используем чанки, чтобы избежать переполнения стека
		const CHUNK_SIZE = 0x8000 // 32KB chunks
		let result = ''

		for (let i = 0; i < data.length; i += CHUNK_SIZE) {
			const chunk = data.subarray(i, Math.min(i + CHUNK_SIZE, data.length))
			result += String.fromCharCode(...chunk)
		}

		return btoa(result)
	}

	// Конвертация base64 в Uint8Array
	private base64ToUint8Array(base64: string): Uint8Array {
		const binaryString = atob(base64)
		const bytes = new Uint8Array(binaryString.length)

		for (let i = 0; i < binaryString.length; i++) {
			bytes[i] = binaryString.charCodeAt(i)
		}

		return bytes
	}

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
		await this.updateUnsyncedCount()

		// Попытаться синхронизировать при инициализации, если есть сеть и сервер доступен
		if (this.canSync()) {
			await this.sync().catch((error) => {
				console.log('Не удалось синхронизировать при инициализации:', error)
			})
		}
	}

	clearStore() {
		// Очищаем реакции
		this.disposeReactions.forEach((dispose) => dispose())
		this.disposeReactions = []

		this.secrets = []
		this.syncStatus = 'idle'
		this.lastSyncTime = null
		this.lastSyncDate = null
		this.unsyncedCount = 0
		this.isLoading = false
		this.isCreating = false
		this.isUpdating = false
		this.isDeleting = false
	}
}
