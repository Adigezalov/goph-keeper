import { observer } from 'mobx-react-lite'

import { StoreContextLogic, TStoreLogic, useStoreLogic } from '@shared/store'

import { HeaderView } from './header.view'

export const Header = observer(() => {
	const store = useStoreLogic<TStoreLogic>(StoreContextLogic)

	const { logout, logoutAll } = store.auth

	const onLogout = () => {
		void logout()
	}

	const onLogoutAll = () => {
		void logoutAll()
	}

	return <HeaderView onLogout={onLogout} onLogoutAll={onLogoutAll} />
})
