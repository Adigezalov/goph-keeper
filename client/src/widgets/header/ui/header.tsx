import { observer } from 'mobx-react-lite'

import { StoreContextLogic, TStoreLogic, useStoreLogic } from '@shared/store'

import { HeaderView } from './header.view'

export const Header = observer(() => {
	const store = useStoreLogic<TStoreLogic>(StoreContextLogic)

	const { logout, logoutAll } = store.auth
	const { isOnline } = store.networkStatus

	const onLogout = () => {
		isOnline ? void logout() : undefined
	}

	const onLogoutAll = () => {
		isOnline ? void logoutAll() : undefined
	}

	return <HeaderView isOnline={isOnline} onLogout={onLogout} onLogoutAll={onLogoutAll} />
})
