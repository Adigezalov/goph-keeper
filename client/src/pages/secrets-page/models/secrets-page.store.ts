import i18next from 'i18next'
import { makeAutoObservable, reaction, runInAction } from 'mobx'
import { v4 as uuidv4 } from 'uuid'
import { StatusCodes } from 'http-status-codes'
import axios from 'axios'

import { db } from '@shared/db'
import { TStoreLogic } from '@shared/store'
import { showToastNotification } from '@shared/toast-notification'
import { TOAST_SEVERITY } from '@shared/uikit/toast'

import {
	createSecretApi,
	deleteSecretApi,
	downloadChunkApi,
	finalizeChunkedUploadApi,
	getSecretApi,
	initChunkedUploadApi,
	syncSecretsApi,
	updateSecretApi,
	uploadChunkApi,
} from '../api'
import {
	TCreateSecretRequest,
	TSecret,
	TSecretForSave,
	TSecretResponse,
	TUpdateSecretRequest,
	TConflictResolution,
} from '../types'
import { mergeChunks, shouldUseChunks, splitIntoChunks } from '../utils'

export class SecretsPageStore {
	secrets: TSecret[] = []

	syncStatus: 'idle' | 'syncing' | 'error' = 'idle'
	lastSyncTime: string | null = null
	lastSyncDate: Date | null = null
	unsyncedCount = 0

	isLoading = false
	isCreating = false
	isUpdating = false
	isDeleting = false

	visibleConflictResolvingModal = false
	conflicts: TConflictResolution[] = []
	currentConflictIndex = -1

	private disposeReactions: (() => void)[] = []

	rootStore: TStoreLogic

	constructor(rootStore: TStoreLogic) {
		this.rootStore = rootStore
		makeAutoObservable(this, {}, { autoBind: true })
		this.setupSyncReactions()
	}

	private setupSyncReactions() {
		const networkReaction = reaction(
			() => this.rootStore.networkStatus.isOnline,
			async (isOnline) => {
				if (isOnline) {
					await this.rootStore.serverStatus.checkStatus()

					if (this.unsyncedCount > 0 && this.rootStore.serverStatus.status) {
						void this.sync()
					}
				}
			},
		)

		const serverReaction = reaction(
			() => this.rootStore.serverStatus.status,
			(status) => {
				if (status && this.rootStore.networkStatus.isOnline && this.unsyncedCount > 0) {
					void this.sync()
				}
			},
		)

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

	setVisibleConflictResolvingModal(visible: boolean) {
		this.visibleConflictResolvingModal = visible
	}

	get currentConflict(): TConflictResolution | null {
		if (this.currentConflictIndex >= 0 && this.currentConflictIndex < this.conflicts.length) {
			return this.conflicts[this.currentConflictIndex]
		}
		return null
	}

	get conflictsCount(): number {
		return this.conflicts.length
	}

	addConflict(conflict: TConflictResolution) {
		const existingIndex = this.conflicts.findIndex(
			(c) => c.secretId === conflict.secretId && c.localId === conflict.localId,
		)

		if (existingIndex === -1) {
			runInAction(() => {
				this.conflicts.push(conflict)
				if (!this.visibleConflictResolvingModal) {
					this.currentConflictIndex = this.conflicts.length - 1
					this.visibleConflictResolvingModal = true
				}
				else if (this.currentConflictIndex === -1) {
					this.currentConflictIndex = 0
				}
			})
		}
	}

	removeCurrentConflict() {
		if (this.currentConflictIndex >= 0 && this.currentConflictIndex < this.conflicts.length) {
			runInAction(() => {
				this.conflicts.splice(this.currentConflictIndex, 1)
				if (this.conflicts.length > 0) {
					if (this.currentConflictIndex >= this.conflicts.length) {
						this.currentConflictIndex = this.conflicts.length - 1
					}
				} else {
					this.currentConflictIndex = -1
					this.visibleConflictResolvingModal = false
				}
			})
		}
	}

	goToNextConflict() {
		if (this.currentConflictIndex < this.conflicts.length - 1) {
			runInAction(() => {
				this.currentConflictIndex++
			})
		}
	}

	goToPrevConflict() {
		if (this.currentConflictIndex > 0) {
			runInAction(() => {
				this.currentConflictIndex--
			})
		}
	}

	get canGoToNext(): boolean {
		return this.currentConflictIndex < this.conflicts.length - 1
	}

	get canGoToPrev(): boolean {
		return this.currentConflictIndex > 0
	}

	async createSecret(data: { secret: TSecretForSave; cb: () => void }) {
		try {
			this.isCreating = true

			const { login, password, metadata, binaryData } = data.secret

			const encryptedLogin = await this.encryptData(login.trim())
			const encryptedPassword = await this.encryptData(password)
			const encryptedBinaryData = binaryData
				? await this.encryptBinaryData(binaryData)
				: undefined

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

			runInAction(() => {
				this.secrets = [newSecret, ...this.secrets]
			})

			data.cb()

			if (this.canSync()) {
				void this.sync()
			}

			await this.updateUnsyncedCount()

			return newSecret
		} catch (error) {
			console.error(i18next.t('secrets.create_error'), error)
			showToastNotification({
				message:
					error instanceof Error ? error.message : i18next.t('secrets.unknown_error'),
				header: i18next.t('secrets.create_error'),
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

			if (this.canSync()) {
				void this.sync()
			}

			await this.updateUnsyncedCount()
		} catch (error) {
			console.error(i18next.t('secrets.update_error'), error)
			showToastNotification({
				message:
					error instanceof Error ? error.message : i18next.t('secrets.unknown_error'),
				header: i18next.t('secrets.update_error'),
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

			if (!existingSecret.id) {
				await db.secrets.delete(id)

				runInAction(() => {
					this.secrets = this.secrets.filter((s) => s.localId !== id)
				})
			} else {
				const deletedSecret: TSecret = {
					...existingSecret,
					syncStatus: 'deleted',
					deletedAt: Date.now(),
					updatedAt: Date.now(),
				}

				await db.secrets.put(deletedSecret)

				runInAction(() => {
					this.secrets = this.secrets.filter((s) => s.localId !== id)
				})

				if (this.canSync()) {
					void this.sync()
				}
			}

			await this.updateUnsyncedCount()
		} catch (error) {
			console.error(i18next.t('secrets.delete_error'), error)
			showToastNotification({
				message:
					error instanceof Error ? error.message : i18next.t('secrets.unknown_error'),
				header: i18next.t('secrets.delete_error'),
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

			const secrets = await db.secrets
				.filter((secret) => !secret.deletedAt && secret.syncStatus !== 'deleted')
				.toArray()

			runInAction(() => {
				this.secrets = secrets
			})
		} catch (error) {
			console.error(i18next.t('secrets.load_error'), error)
		} finally {
			runInAction(() => {
				this.isLoading = false
			})
		}
	}

	private canSync(): boolean {
		return (
			this.rootStore.networkStatus.isOnline &&
			this.rootStore.serverStatus.status &&
			this.syncStatus !== 'syncing'
		)
	}

	async sync() {
		if (!this.canSync()) {
			return
		}

		try {
			runInAction(() => {
				this.syncStatus = 'syncing'
			})

			await this.pushLocalChanges()

			await this.pullServerChanges()

			await this.loadSecrets()

			runInAction(() => {
				this.syncStatus = 'idle'
				this.lastSyncDate = new Date()
			})

			await this.updateUnsyncedCount()

			if (this.conflicts.length > 0 && !this.visibleConflictResolvingModal) {
				runInAction(() => {
					this.currentConflictIndex = 0
					this.visibleConflictResolvingModal = true
				})
			}
		} catch (error) {
			console.error(i18next.t('secrets.sync_error'), error)
			runInAction(() => {
				this.syncStatus = 'error'
			})
		}
	}

	private async pushLocalChanges() {
		const pendingSecrets = await db.secrets
			.filter(
				(secret) => secret.syncStatus === 'pending' || secret.syncStatus === 'deleted',
			)
			.toArray()

		for (const secret of pendingSecrets) {
			try {
				if (secret.syncStatus === 'deleted' && secret.id) {
					await deleteSecretApi(secret.id)
					await db.secrets.delete(secret.localId)
				} else if (!secret.id) {
					if (secret.binaryData && shouldUseChunks(secret.binaryData.length)) {
						const serverSecret = await this.uploadWithChunks(
							secret.binaryData,
							secret.login,
							secret.password,
							secret.metadata,
						)
						await db.secrets.put({
							...secret,
							id: serverSecret.id,
							version: serverSecret.version,
							syncStatus: 'synced',
						})
					} else {
						const response = await createSecretApi(this.secretToCreateRequest(secret))
						await db.secrets.put({
							...secret,
							id: response.data.id,
							version: response.data.version,
							syncStatus: 'synced',
						})
					}
				} else {
					if (secret.binaryData && shouldUseChunks(secret.binaryData.length)) {
						try {
							const serverSecret = await this.uploadWithChunks(
								secret.binaryData,
								secret.login,
								secret.password,
								secret.metadata,
								secret.version,
							)
							await db.secrets.put({
								...secret,
								version: serverSecret.version,
								syncStatus: 'synced',
							})
						} catch (error) {
							if (axios.isAxiosError(error) && error.response?.status === StatusCodes.CONFLICT) {
								await this.handleVersionConflict(secret, 'sync')
							} else {
								console.error(i18next.t('secrets.send_error'), error)
							}
						}
					} else {
						try {
							const response = await updateSecretApi(
								secret.id,
								this.secretToUpdateRequest(secret),
							)
							await db.secrets.put({
								...secret,
								version: response.data.version,
								syncStatus: 'synced',
							})
						} catch (error) {
							if (axios.isAxiosError(error) && error.response?.status === StatusCodes.CONFLICT) {
								await this.handleVersionConflict(secret, 'sync')
							} else {
								console.error(i18next.t('secrets.send_error'), error)
							}
						}
					}
				}
			} catch (error) {
				if (axios.isAxiosError(error) && error.response?.status === StatusCodes.CONFLICT) {
					// Conflict handled above
				} else {
					console.error(i18next.t('secrets.send_error'), error)
				}
			}
		}
	}

	private async pullServerChanges() {
		try {
			const response = await syncSecretsApi(this.lastSyncTime || undefined)

			for (const serverSecret of response.data.secrets) {
				await this.applyServerSecret(serverSecret)
			}

			runInAction(() => {
				this.lastSyncTime = response.data.server_time
			})

			await this.saveLastSyncTime(response.data.server_time)
		} catch (error) {
			console.error(i18next.t('secrets.receive_error'), error)
			throw error
		}
	}

	private async applyServerSecret(serverSecret: TSecretResponse) {
		const existingSecret = await db.secrets.where('id').equals(serverSecret.id).first()

		if (serverSecret.deleted_at) {
			if (existingSecret) {
				if (existingSecret.syncStatus === 'pending') {
					await this.handleVersionConflict(existingSecret, 'sync')
					return
				}
				await db.secrets.delete(existingSecret.localId)
			}
			return
		}

		if (existingSecret) {
			if (
				existingSecret.syncStatus === 'pending' &&
				serverSecret.version > existingSecret.version
			) {
				await this.handleVersionConflict(existingSecret, 'sync')
				return
			}

			if (
				existingSecret.version === serverSecret.version &&
				existingSecret.syncStatus === 'synced'
			) {
				return
			}

			if (existingSecret.syncStatus === 'synced' || serverSecret.version > existingSecret.version) {
				let binaryData: Uint8Array | undefined = existingSecret.binaryData
				if (serverSecret.binary_data) {
					binaryData = this.base64ToUint8Array(serverSecret.binary_data)
				} else if (serverSecret.binary_data_size && serverSecret.binary_data_size > 0) {
					binaryData = await this.downloadWithChunks(serverSecret.id)
				}

				await db.secrets.put({
					localId: existingSecret.localId,
					id: existingSecret.id,
					login: serverSecret.login,
					password: serverSecret.password,
					metadata: serverSecret.metadata as Record<string, string> | undefined,
					binaryData,
					version: serverSecret.version,
					syncStatus: 'synced',
					createdAt: existingSecret.createdAt,
					updatedAt: new Date(serverSecret.updated_at).getTime(),
					deletedAt: existingSecret.deletedAt,
				})
			}
		} else {
			let binaryData: Uint8Array | undefined
			if (serverSecret.binary_data) {
				binaryData = this.base64ToUint8Array(serverSecret.binary_data)
			} else if (serverSecret.binary_data_size && serverSecret.binary_data_size > 0) {
				binaryData = await this.downloadWithChunks(serverSecret.id)
			}

			const newSecret: TSecret = {
				localId: uuidv4(),
				id: serverSecret.id,
				login: serverSecret.login,
				password: serverSecret.password,
				metadata: serverSecret.metadata as Record<string, string> | undefined,
				binaryData,
				version: serverSecret.version,
				syncStatus: 'synced',
				createdAt: new Date(serverSecret.created_at).getTime(),
				updatedAt: new Date(serverSecret.updated_at).getTime(),
			}
			await db.secrets.add(newSecret)
		}
	}

	async updateUnsyncedCount() {
		const count = await db.secrets
			.filter(
				(secret) => secret.syncStatus === 'pending' || secret.syncStatus === 'deleted',
			)
			.count()
		runInAction(() => {
			this.unsyncedCount = count
		})
	}

	private async loadLastSyncTime() {
		try {
			const meta = await db.syncMeta.get('lastSyncTime')
			if (meta) {
				runInAction(() => {
					this.lastSyncTime = meta.value
				})
			}
		} catch (error) {
			console.error(i18next.t('secrets.load_sync_time_error'), error)
		}
	}

	private async saveLastSyncTime(time: string) {
		try {
			await db.syncMeta.put({ key: 'lastSyncTime', value: time })
		} catch (error) {
			console.error(i18next.t('secrets.save_sync_time_error'), error)
		}
	}

	private secretToCreateRequest(secret: TSecret): TCreateSecretRequest {
		return {
			login: secret.login,
			password: secret.password,
			metadata: secret.metadata,
			binary_data: secret.binaryData
				? this.uint8ArrayToBase64(secret.binaryData)
				: undefined,
		}
	}

	private secretToUpdateRequest(secret: TSecret): TUpdateSecretRequest {
		return {
			login: secret.login,
			password: secret.password,
			metadata: secret.metadata,
			binary_data: secret.binaryData
				? this.uint8ArrayToBase64(secret.binaryData)
				: undefined,
			version: secret.version,
		}
	}

	private uint8ArrayToBase64(data: Uint8Array): string {
		const CHUNK_SIZE = 0x8000
		let result = ''

		for (let i = 0; i < data.length; i += CHUNK_SIZE) {
			const chunk = data.subarray(i, Math.min(i + CHUNK_SIZE, data.length))
			result += String.fromCharCode(...chunk)
		}

		return btoa(result)
	}

	private base64ToUint8Array(base64: string): Uint8Array {
		const binaryString = atob(base64)
		const bytes = new Uint8Array(binaryString.length)

		for (let i = 0; i < binaryString.length; i++) {
			bytes[i] = binaryString.charCodeAt(i)
		}

		return bytes
	}

	private async uploadWithChunks(
		data: Uint8Array,
		login: string,
		password: string,
		metadata?: Record<string, string>,
		version?: number,
	): Promise<TSecretResponse> {
		const chunks = splitIntoChunks(data)
		const totalChunks = chunks.length

		const initResponse = await initChunkedUploadApi({
			totalChunks,
			totalSize: data.length,
			metadata,
		})

		const { uploadId, secretId } = initResponse.data

		for (let i = 0; i < chunks.length; i++) {
			const chunk = chunks[i]
			const chunkBase64 = this.uint8ArrayToBase64(chunk)

			await uploadChunkApi(secretId, {
				uploadId,
				chunkIndex: i,
				totalChunks,
				data: chunkBase64,
			})
		}

		const finalizeResponse = await finalizeChunkedUploadApi(secretId, {
			uploadId,
			login,
			password,
			metadata,
			version,
		})

		return finalizeResponse.data
	}

	private async downloadWithChunks(secretId: string): Promise<Uint8Array> {
		const firstChunk = await downloadChunkApi(secretId, 0)
		const { totalChunks } = firstChunk.data

		const chunks: Uint8Array[] = []
		chunks[0] = this.base64ToUint8Array(firstChunk.data.data)

		for (let i = 1; i < totalChunks; i++) {
			const chunkResponse = await downloadChunkApi(secretId, i)
			chunks[i] = this.base64ToUint8Array(chunkResponse.data.data)
		}

		return mergeChunks(chunks)
	}

	private async encryptData(data: string): Promise<string> {
		const cryptoKey = this.rootStore.cryptoKey.cryptoKey

		if (!cryptoKey) {
			throw new Error(i18next.t('crypto.key_not_set'))
		}

		const iv = window.crypto.getRandomValues(new Uint8Array(12))

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

		const combined = new Uint8Array(iv.length + encrypted.byteLength)
		combined.set(iv, 0)
		combined.set(new Uint8Array(encrypted), iv.length)

		return btoa(String.fromCharCode(...combined))
	}

	async decryptData(encryptedData: string): Promise<string> {
		const cryptoKey = this.rootStore.cryptoKey.cryptoKey

		if (!cryptoKey) {
			throw new Error(i18next.t('crypto.key_not_set'))
		}

		const combined = Uint8Array.from(atob(encryptedData), (c) => c.charCodeAt(0))

		const iv = combined.slice(0, 12)
		const encrypted = combined.slice(12)

		const decrypted = await window.crypto.subtle.decrypt(
			{
				name: 'AES-GCM',
				iv,
			},
			cryptoKey,
			encrypted,
		)

		const decoder = new TextDecoder()
		return decoder.decode(decrypted)
	}

	private async encryptBinaryData(data: Uint8Array): Promise<Uint8Array> {
		const cryptoKey = this.rootStore.cryptoKey.cryptoKey

		if (!cryptoKey) {
			throw new Error(i18next.t('crypto.key_not_set'))
		}

		const iv = window.crypto.getRandomValues(new Uint8Array(12))

		const encrypted = await window.crypto.subtle.encrypt(
			{
				name: 'AES-GCM',
				iv,
			},
			cryptoKey,
			data.buffer as ArrayBuffer,
		)

		const combined = new Uint8Array(iv.length + encrypted.byteLength)
		combined.set(iv, 0)
		combined.set(new Uint8Array(encrypted), iv.length)

		return combined
	}

	async decryptBinaryData(encryptedData: Uint8Array): Promise<Uint8Array> {
		const cryptoKey = this.rootStore.cryptoKey.cryptoKey

		if (!cryptoKey) {
			throw new Error(i18next.t('crypto.key_not_set'))
		}

		const iv = encryptedData.slice(0, 12)
		const encrypted = encryptedData.slice(12)

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

		await this.loadLastSyncTime()

		if (this.canSync()) {
			await this.sync().catch((error) => {
				console.log(i18next.t('secrets.sync_init_error'), error)
			})
		}
	}

	private async handleVersionConflict(secret: TSecret, conflictType: 'update' | 'sync') {
		if (!secret.id) {
			console.error('Cannot handle conflict for secret without id')
			return
		}

		try {
			const serverResponse = await getSecretApi(secret.id)
			const serverSecret = serverResponse.data

			let serverBinaryData: Uint8Array | undefined
			if (serverSecret.binary_data) {
				serverBinaryData = this.base64ToUint8Array(serverSecret.binary_data)
			} else if (serverSecret.binary_data_size && serverSecret.binary_data_size > 0) {
				serverBinaryData = await this.downloadWithChunks(serverSecret.id)
			}

			const conflict: TConflictResolution = {
				secretId: secret.id,
				localId: secret.localId,
				localVersion: secret,
				serverVersion: serverSecret,
				serverBinaryData,
				conflictType,
				timestamp: Date.now(),
			}

			runInAction(() => {
				this.addConflict(conflict)
			})
		} catch (error) {
			console.error('Error loading server version for conflict resolution:', error)
			showToastNotification({
				message: i18next.t('secrets.conflict_load_error'),
				header: i18next.t('secrets.conflict_error'),
				severity: TOAST_SEVERITY.ERROR,
			})
		}
	}

	async resolveConflict(choice: 'local' | 'server') {
		const conflict = this.currentConflict
		if (!conflict) {
			return
		}

		try {
			const maxVersion = Math.max(
				conflict.localVersion.version,
				conflict.serverVersion.version,
			)

			if (choice === 'local') {
				const localSecret = conflict.localVersion
				const hasBinaryData = localSecret.binaryData && localSecret.binaryData.length > 0

				if (hasBinaryData && shouldUseChunks(localSecret.binaryData!.length)) {
					const serverSecret = await this.uploadWithChunks(
						localSecret.binaryData!,
						localSecret.login,
						localSecret.password,
						localSecret.metadata,
						maxVersion,
					)
					const updatedSecret: TSecret = {
						localId: localSecret.localId,
						id: localSecret.id,
						login: localSecret.login,
						password: localSecret.password,
						metadata: localSecret.metadata ? { ...localSecret.metadata } : undefined,
						binaryData: localSecret.binaryData,
						version: serverSecret.version,
						syncStatus: 'synced',
						createdAt: localSecret.createdAt,
						updatedAt: Date.now(),
						deletedAt: localSecret.deletedAt,
					}
					await db.secrets.put(updatedSecret)
				} else {
					// Обычное обновление
					const requestData: TUpdateSecretRequest = {
						login: localSecret.login,
						password: localSecret.password,
						metadata: localSecret.metadata || {},
						binary_data: localSecret.binaryData
							? this.uint8ArrayToBase64(localSecret.binaryData)
							: undefined,
						version: maxVersion,
					}

					const response = await updateSecretApi(conflict.secretId, requestData)

					// Обновляем локальную БД
					const updatedSecret: TSecret = {
						localId: localSecret.localId,
						id: localSecret.id,
						login: localSecret.login,
						password: localSecret.password,
						metadata: localSecret.metadata ? { ...localSecret.metadata } : undefined,
						binaryData: localSecret.binaryData,
						version: response.data.version,
						syncStatus: 'synced',
						createdAt: localSecret.createdAt,
						updatedAt: Date.now(),
						deletedAt: localSecret.deletedAt,
					}

					await db.secrets.put(updatedSecret)
				}
			} else {
				const serverSecret = conflict.serverVersion
				const serverBinaryData = conflict.serverBinaryData
				const localSecret = conflict.localVersion
				const hasBinaryData = serverBinaryData && serverBinaryData.length > 0

				if (hasBinaryData && shouldUseChunks(serverBinaryData!.length)) {
					const updatedSecretResponse = await this.uploadWithChunks(
						serverBinaryData!,
						serverSecret.login,
						serverSecret.password,
						serverSecret.metadata as Record<string, string> | undefined,
						maxVersion,
					)
					const updatedSecret: TSecret = {
						localId: localSecret.localId,
						id: localSecret.id,
						login: serverSecret.login,
						password: serverSecret.password,
						metadata: serverSecret.metadata
							? ({ ...serverSecret.metadata } as Record<string, string>)
							: undefined,
						binaryData: serverBinaryData,
						version: updatedSecretResponse.version,
						syncStatus: 'synced',
						createdAt: localSecret.createdAt,
						updatedAt: Date.now(),
						deletedAt: localSecret.deletedAt,
					}
					await db.secrets.put(updatedSecret)
				} else {
					// Обычное обновление
					const requestData: TUpdateSecretRequest = {
						login: serverSecret.login,
						password: serverSecret.password,
						metadata: (serverSecret.metadata as Record<string, string>) || {},
						binary_data: serverSecret.binary_data,
						version: maxVersion,
					}

					const response = await updateSecretApi(conflict.secretId, requestData)

					// Обновляем локальную БД
					const updatedSecret: TSecret = {
						localId: localSecret.localId,
						id: localSecret.id,
						login: serverSecret.login,
						password: serverSecret.password,
						metadata: serverSecret.metadata
							? ({ ...serverSecret.metadata } as Record<string, string>)
							: undefined,
						binaryData: serverBinaryData,
						version: response.data.version,
						syncStatus: 'synced',
						createdAt: localSecret.createdAt,
						updatedAt: Date.now(),
						deletedAt: localSecret.deletedAt,
					}

					await db.secrets.put(updatedSecret)
				}
			}

			await this.loadSecrets()
			await this.updateUnsyncedCount()

			const wasLastConflict = this.conflicts.length === 1

			this.removeCurrentConflict()

			if (wasLastConflict) {
				showToastNotification({
					message: i18next.t('secrets.all_conflicts_resolved'),
					header: i18next.t('secrets.success'),
					severity: TOAST_SEVERITY.SUCCESS,
				})
			} else {
				showToastNotification({
					message: i18next.t('secrets.conflict_resolved'),
					header: i18next.t('secrets.success'),
					severity: TOAST_SEVERITY.SUCCESS,
				})
			}
		} catch (error) {
			console.error('Error resolving conflict:', error)
			
			const isServerError = axios.isAxiosError(error) && error.response
			
			if (!isServerError) {
				this.removeCurrentConflict()
				
				if (this.canSync()) {
					void this.sync()
				}
			}
			
			showToastNotification({
				message:
					error instanceof Error ? error.message : i18next.t('secrets.conflict_resolve_error'),
				header: i18next.t('secrets.conflict_error'),
				severity: TOAST_SEVERITY.ERROR,
			})
		}
	}

	clearSecretsPageStore() {
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

		this.visibleConflictResolvingModal = false
		this.conflicts = []
		this.currentConflictIndex = -1
	}

	openConflictResolutionModal() {
		if (this.conflicts.length > 0) {
			runInAction(() => {
				if (this.currentConflictIndex === -1) {
					this.currentConflictIndex = 0
				}
				this.visibleConflictResolvingModal = true
			})
		}
	}
}
