import { ReactElement } from 'react'

export type TBook = {
	name: string
	path: string
	page?: ReactElement
	icon?: (x?: any) => ReactElement
	externalUrl?: string
	isRoot: boolean
	isShow: boolean
	isMenu: boolean
	permissions?: []
	pages?: TBook[]
}
