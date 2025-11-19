import { observer } from 'mobx-react-lite'
import { ReactNode, useEffect, useRef } from 'react'

import {
	MAX_RECONNECT_ATTEMPTS,
	REALTIME_WS_URL,
	RECONNECT_BASE_DELAY_MS,
	RECONNECT_MAX_DELAY_MS,
	SYNC_DEBOUNCE_MS,
	type SecretEventMessage,
} from '@entities/realtime'

import { StoreContextLogic, TStoreLogic, useStoreLogic } from '@shared/store'
import { getAccessToken } from '@shared/tokens'

type TProps = {
	children: ReactNode
}

export const RealtimeProvider = observer(({ children }: TProps) => {
	const store = useStoreLogic<TStoreLogic>(StoreContextLogic)
	const wsRef = useRef<WebSocket | null>(null)
	const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null)
	const syncTimeoutRef = useRef<NodeJS.Timeout | null>(null)

	const { auth, realtime, secretsPage } = store

	// Debounced синхронизация
	const scheduleSync = () => {
		if (syncTimeoutRef.current) {
			clearTimeout(syncTimeoutRef.current)
		}

		syncTimeoutRef.current = setTimeout(() => {
			// sync() сам проверяет canSync() внутри
			void secretsPage.sync()
		}, SYNC_DEBOUNCE_MS)
	}

	// Обработка сообщений от WebSocket
	const handleMessage = (event: MessageEvent) => {
		try {
			const message: SecretEventMessage = JSON.parse(event.data)

			console.log(message)

			// Проверяем, что сообщение для нашего пользователя
			// (на сервере уже проверяется, но для безопасности проверяем и здесь)

			// Запускаем синхронизацию с debounce
			scheduleSync()
		} catch (error) {
			console.error('[Realtime] Ошибка парсинга сообщения:', error)
		}
	}

	// Обработка открытия соединения
	const handleOpen = () => {
		console.log('[Realtime] WebSocket соединение установлено')
		realtime.setConnectionStatus('connected')
		realtime.resetReconnectAttempts()
	}

	// Обработка закрытия соединения
	const handleClose = (event: CloseEvent) => {
		console.log('[Realtime] WebSocket соединение закрыто:', event.code, event.reason)
		realtime.setConnectionStatus('disconnected')
		wsRef.current = null

		// Переподключаемся только если пользователь авторизован
		if (auth.auth && realtime.reconnectAttempts < MAX_RECONNECT_ATTEMPTS) {
			reconnect()
		}
	}

	// Обработка ошибок
	const handleError = (error: Event) => {
		console.error('[Realtime] WebSocket ошибка:', error)
		realtime.setConnectionStatus('disconnected')
	}

	// Подключение к WebSocket
	const connect = () => {
		if (wsRef.current?.readyState === WebSocket.OPEN) {
			return // Уже подключено
		}

		const token = getAccessToken()
		if (!token) {
			console.warn('[Realtime] Нет токена для подключения')
			return
		}

		// Генерируем sessionID если его нет
		if (!realtime.sessionID) {
			realtime.generateSessionID()
		}

		const wsUrl = REALTIME_WS_URL(token, realtime.sessionID || undefined)

		console.log('[Realtime] Подключение к WebSocket:', wsUrl.replace(token, 'TOKEN'))

		realtime.setConnectionStatus('connecting')

		try {
			const ws = new WebSocket(wsUrl)
			wsRef.current = ws

			ws.onopen = handleOpen
			ws.onmessage = handleMessage
			ws.onclose = handleClose
			ws.onerror = handleError
		} catch (error) {
			console.error('[Realtime] Ошибка создания WebSocket:', error)
			realtime.setConnectionStatus('disconnected')
		}
	}

	// Переподключение с экспоненциальной задержкой
	const reconnect = () => {
		if (reconnectTimeoutRef.current) {
			clearTimeout(reconnectTimeoutRef.current)
		}

		const delay = Math.min(
			RECONNECT_BASE_DELAY_MS * Math.pow(2, realtime.reconnectAttempts),
			RECONNECT_MAX_DELAY_MS,
		)

		console.log(
			`[Realtime] Переподключение через ${delay}ms (попытка ${realtime.reconnectAttempts + 1})`,
		)

		realtime.setConnectionStatus('reconnecting')
		realtime.incrementReconnectAttempts()

		reconnectTimeoutRef.current = setTimeout(() => {
			connect()
		}, delay)
	}

	// Отключение от WebSocket
	const disconnect = () => {
		if (reconnectTimeoutRef.current) {
			clearTimeout(reconnectTimeoutRef.current)
			reconnectTimeoutRef.current = null
		}

		if (syncTimeoutRef.current) {
			clearTimeout(syncTimeoutRef.current)
			syncTimeoutRef.current = null
		}

		if (wsRef.current) {
			wsRef.current.close()
			wsRef.current = null
		}

		realtime.setConnectionStatus('disconnected')
	}

	// Подключаемся при авторизации
	useEffect(() => {
		if (auth.auth) {
			connect()
		} else {
			disconnect()
			realtime.clearRealtimeStore()
		}

		return () => {
			disconnect()
		}
	}, [auth.auth])

	// Очистка при размонтировании
	useEffect(() => {
		return () => {
			disconnect()
		}
	}, [])

	return <>{children}</>
})
