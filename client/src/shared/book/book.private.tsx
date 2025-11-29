import { TBook } from './book.types'
import { SecretsPage } from './lazy-pages'

export const bookPrivate: TBook[] = [
	{
		name: 'secrets',
		path: '/secrets',
		page: <SecretsPage />,
		isRoot: true,
		isShow: true,
		isMenu: true,
	},
]
