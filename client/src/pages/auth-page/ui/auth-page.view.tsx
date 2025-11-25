import { Control } from 'react-hook-form'
import { useTranslation } from 'react-i18next'

import { TAuth } from '@entities/auth'

import { InputPasswordField } from '@shared/reused/input-password-field'
import { InputTextField } from '@shared/reused/input-text-field'
import { Button } from '@shared/uikit/button'

import styles from './auth-page.module.sass'

type Props = {
	control: Control<TAuth>
	title: string
	buttonTitle: string
	loading: boolean
	onLogin: () => void
	redirect: {
		label: string
		onClick: () => void
	}
}

export const AuthPageView = ({
	control,
	title,
	buttonTitle,
	loading,
	onLogin,
	redirect,
}: Props) => {
	const { t } = useTranslation()

	return (
		<div className={styles.root}>
			<div className={styles.form}>
				<span className={styles.form_title}>{title}</span>
				<form className={styles.form_content} onSubmit={onLogin} noValidate>
					<div className={styles.form_data}>
						<InputTextField<TAuth>
							control={control}
							name="email"
							label={t('email')}
							rules={{
								validate: (value) => {
									if (value && value.trim() !== '') {
										const emailRegex = /^[A-Z0-9._%+-]+@[A-Z0-9.-]+\.[A-Z]{2,}$/i
										return (
											emailRegex.test(value) || t('validation_error.invalid_email_format')
										)
									}
									return true
								},
							}}
							required
							disabled={loading}
						/>
						<InputPasswordField<TAuth>
							control={control}
							name="password"
							label={t('password')}
							required
							disabled={loading}
						/>
					</div>
					<div className={styles.actions}>
						<Button label={buttonTitle} type={'submit'} loading={loading} />
						<div className={styles.redirect}>
							<Button
								label={redirect.label}
								link
								onClick={redirect.onClick}
								type={'button'}
							/>
						</div>
					</div>
				</form>
			</div>
		</div>
	)
}
