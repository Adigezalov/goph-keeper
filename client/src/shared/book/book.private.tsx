import { TBook } from './book.types'
import { DashboardPage } from './lazy-pages'

export const bookPrivate: TBook[] = [
	{
		name: 'dashboard',
		path: '/dashboard',
		page: <DashboardPage />,
		isRoot: true,
		isShow: true,
		isMenu: true,
	},
]
