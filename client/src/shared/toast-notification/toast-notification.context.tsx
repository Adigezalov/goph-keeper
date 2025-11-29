import { createContext } from 'react'

import { TToastNotificationContext } from './toast-notification.types'

export const ToastNotificationContext = createContext<TToastNotificationContext>({
	show: () => null,
})
