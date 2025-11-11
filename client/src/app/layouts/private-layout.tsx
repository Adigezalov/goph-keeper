import { Outlet } from 'react-router'

import { Header } from '@widgets/header'

import { AuthProvider } from '../providers'

import styles from './layout.module.sass'

export const PrivateLayout = () => {
	return (
		<AuthProvider>
			<div className={styles.root}>
				<Header />
				<div className={styles.content}>
					<Outlet />
				</div>
			</div>
		</AuthProvider>
	)
}
