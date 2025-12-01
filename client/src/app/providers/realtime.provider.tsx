import { observer } from 'mobx-react-lite'
import { ReactNode, useEffect, useRef } from 'react'
import { useTranslation } from 'react-i18next'

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
	const { t } = useTranslation()

	const wsRef = useRef<WebSocket | null>(null)
	const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null)
	const syncTimeoutRef = useRef<NodeJS.Timeout | null>(null)

	const store = useStoreLogic<TStoreLogic>(StoreContextLogic)

	const { auth, realtime, secretsPage } = store

	const scheduleSync = () => {
		if (syncTimeoutRef.current) {
			clearTimeout(syncTimeoutRef.current)
		}

		syncTimeoutRef.current = setTimeout(() => {
			void secretsPage.sync()
		}, SYNC_DEBOUNCE_MS)
	}

	const handleMessage = (event: MessageEvent) => {
		try {
			const message: SecretEventMessage = JSON.parse(event.data)

			console.log(message)

			scheduleSync()
		} catch (error) {
			console.error(t('realtime.error_parsing_message'), error)
		}
	}

	const handleOpen = () => {
		console.log(t('realtime.ws_connect_success'))
		realtime.setConnectionStatus('connected')
		realtime.resetReconnectAttempts()
	}

	const handleClose = (event: CloseEvent) => {
		console.log(t('realtime.ws_connect_close'), event.code, event.reason)
		realtime.setConnectionStatus('disconnected')
		wsRef.current = null

		if (auth.auth && realtime.reconnectAttempts < MAX_RECONNECT_ATTEMPTS) {
			reconnect()
		}
	}

	const handleError = (error: Event) => {
		console.error(t('realtime.error_ws'), error)
		realtime.setConnectionStatus('disconnected')
	}

	const connect = () => {
		if (wsRef.current?.readyState === WebSocket.OPEN) {
			return
		}

		const token = getAccessToken()
		if (!token) {
			console.warn(t('realtime.not_token_to_connect'))
			return
		}

		if (!realtime.sessionID) {
			realtime.generateSessionID()
		}

		const wsUrl = REALTIME_WS_URL(token, realtime.sessionID || undefined)

		console.log(t('realtime.connect_to_ws'), wsUrl.replace(token, 'TOKEN'))

		realtime.setConnectionStatus('connecting')

		try {
			const ws = new WebSocket(wsUrl)
			wsRef.current = ws

			ws.onopen = handleOpen
			ws.onmessage = handleMessage
			ws.onclose = handleClose
			ws.onerror = handleError
		} catch (error) {
			console.error(t('realtime.error_create_ws'), error)
			realtime.setConnectionStatus('disconnected')
		}
	}

	const reconnect = () => {
		if (reconnectTimeoutRef.current) {
			clearTimeout(reconnectTimeoutRef.current)
		}

		const delay = Math.min(
			RECONNECT_BASE_DELAY_MS * Math.pow(2, realtime.reconnectAttempts),
			RECONNECT_MAX_DELAY_MS,
		)

		console.log(
			t('realtime.reconnect_through', {
				delay,
				attempt: realtime.reconnectAttempts + 1,
			}),
		)

		realtime.setConnectionStatus('reconnecting')
		realtime.incrementReconnectAttempts()

		reconnectTimeoutRef.current = setTimeout(() => {
			connect()
		}, delay)
	}

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
		// eslint-disable-next-line react-hooks/exhaustive-deps
	}, [auth.auth])

	useEffect(() => {
		return () => {
			disconnect()
		}
		// eslint-disable-next-line react-hooks/exhaustive-deps
	}, [])

	return <>{children}</>
})
