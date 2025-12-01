import { useContext } from 'react'

import { ToastNotificationContext } from './toast-notification.context'
import { type TToastNotificationContext } from './toast-notification.types'

export const useToastNotification = (): TToastNotificationContext =>
	useContext(ToastNotificationContext)
