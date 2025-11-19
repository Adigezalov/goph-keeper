import { makeAutoObservable, runInAction } from 'mobx'
import { v4 as uuidv4 } from 'uuid'

import { TStoreLogic } from '@shared/store'

import type { RealtimeConnectionStatus } from '../types'

export class RealtimeStore {
	sessionID: string | null = null
	connectionStatus: RealtimeConnectionStatus = 'disconnected'
	reconnectAttempts = 0
	lastReconnectTime: Date | null = null

	rootStore: TStoreLogic

	constructor(rootStore: TStoreLogic) {
		this.rootStore = rootStore
		makeAutoObservable(this, {}, { autoBind: true })
		this.loadSessionID()
	}

	// Загружает sessionID из localStorage
	private loadSessionID() {
		const stored = localStorage.getItem('realtime_session_id')
		if (stored) {
			this.sessionID = stored
		}
	}

	// Генерирует новый sessionID
	generateSessionID() {
		const newSessionID = uuidv4()
		this.setSessionID(newSessionID)
		return newSessionID
	}

	// Устанавливает sessionID
	setSessionID(id: string) {
		runInAction(() => {
			this.sessionID = id
			localStorage.setItem('realtime_session_id', id)
		})
	}

	// Очищает sessionID
	clearSessionID() {
		runInAction(() => {
			this.sessionID = null
			localStorage.removeItem('realtime_session_id')
		})
	}

	// Устанавливает статус подключения
	setConnectionStatus(status: RealtimeConnectionStatus) {
		runInAction(() => {
			this.connectionStatus = status
		})
	}

	// Увеличивает счетчик попыток переподключения
	incrementReconnectAttempts() {
		runInAction(() => {
			this.reconnectAttempts += 1
			this.lastReconnectTime = new Date()
		})
	}

	// Сбрасывает счетчик попыток переподключения
	resetReconnectAttempts() {
		runInAction(() => {
			this.reconnectAttempts = 0
			this.lastReconnectTime = null
		})
	}

	// Очищает store при выходе
	clearRealtimeStore() {
		this.clearSessionID()
		this.setConnectionStatus('disconnected')
		this.resetReconnectAttempts()
	}
}

