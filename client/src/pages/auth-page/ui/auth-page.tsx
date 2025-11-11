import { observer } from 'mobx-react-lite'
import { useMemo, useState } from 'react'
import { useForm } from 'react-hook-form'
import { useTranslation } from 'react-i18next'

import { TAuth } from '@entities/auth'

import { StoreContextLogic, TStoreLogic, useStoreLogic } from '@shared/store'

import { AuthPageView } from './auth-page.view'

export const AuthPage = observer(() => {
	const { t } = useTranslation()

	const [isRegistration, setIsRegistration] = useState<boolean>(false)

	const store = useStoreLogic<TStoreLogic>(StoreContextLogic)

	const { isLoginLoading, login, registration } = store.auth

	const { control, handleSubmit } = useForm<TAuth>({
		mode: 'all',
		defaultValues: {
			email: '',
			password: '',
		},
	})

	const title = useMemo(() => {
		return t(isRegistration ? 'register' : 'login_to_system')
	}, [isRegistration, t])

	const buttonTitle = useMemo(() => {
		return t(isRegistration ? 'register' : 'login')
	}, [isRegistration, t])

	const redirectTitle = useMemo(() => {
		return t(isRegistration ? 'login' : 'register')
	}, [isRegistration, t])

	const onSubmit = (data: TAuth) => {
		isRegistration ? void registration(data) : void login(data)
	}

	const onLogin = handleSubmit(onSubmit)

	const onRedirect = () => {
		setIsRegistration(!isRegistration)
	}

	return (
		<AuthPageView
			control={control}
			title={title}
			buttonTitle={buttonTitle}
			onLogin={onLogin}
			loading={isLoginLoading}
			redirect={{ label: redirectTitle, onClick: onRedirect }}
		/>
	)
})
