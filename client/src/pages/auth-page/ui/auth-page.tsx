import { observer } from 'mobx-react-lite'
import { useMemo, useState } from 'react'
import { useForm } from 'react-hook-form'
import { useTranslation } from 'react-i18next'

import { TAuth } from '@entities/auth'

import { StoreContextLogic, TStoreLogic, useStoreLogic } from '@shared/store'

import { AuthPageView } from './auth-page.view'
import { VerifyEmailView } from './verify-email.view'

type TVerificationForm = {
	code: string
}

export const AuthPage = observer(() => {
	const { t } = useTranslation()

	const [isRegistration, setIsRegistration] = useState<boolean>(false)

	const store = useStoreLogic<TStoreLogic>(StoreContextLogic)

	const { 
		isLoginLoading, 
		isRegistrationLoading, 
		isVerifyEmailLoading, 
		isResendCodeLoading,
		awaitingEmailVerification,
		verificationEmail,
		login, 
		registration,
		verifyEmail,
		resendCode
	} = store.auth

	const { control, handleSubmit } = useForm<TAuth>({
		mode: 'all',
		defaultValues: {
			email: '',
			password: '',
		},
	})

	const { control: verifyControl, handleSubmit: handleVerifySubmit } = useForm<TVerificationForm>({
		mode: 'all',
		defaultValues: {
			code: '',
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

	const onVerifySubmit = (data: TVerificationForm) => {
		void verifyEmail({ email: verificationEmail, code: data.code })
	}

	const onVerify = handleVerifySubmit(onVerifySubmit)

	const onResend = () => {
		void resendCode({ email: verificationEmail })
	}

	if (awaitingEmailVerification) {
		return (
			<VerifyEmailView
				control={verifyControl}
				email={verificationEmail}
				loading={isVerifyEmailLoading}
				resendLoading={isResendCodeLoading}
				onVerify={onVerify}
				onResend={onResend}
			/>
		)
	}

	return (
		<AuthPageView
			control={control}
			title={title}
			buttonTitle={buttonTitle}
			onLogin={onLogin}
			loading={isLoginLoading || isRegistrationLoading}
			redirect={{ label: redirectTitle, onClick: onRedirect }}
		/>
	)
})
