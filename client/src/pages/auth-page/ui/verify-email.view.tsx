import { Control } from 'react-hook-form'
import { useTranslation } from 'react-i18next'

import { InputTextField } from '@shared/reused/input-text-field'
import { Button } from '@shared/uikit/button'

import styles from './auth-page.module.sass'

type TVerificationForm = {
	code: string
}

type Props = {
	control: Control<TVerificationForm>
	email: string
	loading: boolean
	resendLoading: boolean
	onVerify: () => void
	onResend: () => void
}

export const VerifyEmailView = ({
	control,
	email,
	loading,
	resendLoading,
	onVerify,
	onResend,
}: Props) => {
	const { t } = useTranslation()

	return (
		<div className={styles.root}>
			<div className={styles.form}>
				<span className={styles.form_title}>{t('verify_email')}</span>
				<p className={styles.form_description}>
					{t('verification_code_sent_to')} {email}
				</p>
				<form className={styles.form_content} onSubmit={onVerify} noValidate>
					<div className={styles.form_data}>
						<InputTextField<TVerificationForm>
							control={control}
							name="code"
							label={t('verification_code')}
							rules={{
								validate: (value) => {
									if (value && value.trim() !== '') {
										const codeRegex = /^\d{6}$/
										return codeRegex.test(value) || t('validation_error.invalid_code_format')
									}
									return true
								},
							}}
							required
							disabled={loading}
							placeholder="123456"
						/>
					</div>
					<div className={styles.actions}>
						<Button label={t('verify')} type={'submit'} loading={loading} />
						<div className={styles.redirect}>
							<Button
								label={t('resend_code')}
								link
								onClick={onResend}
								type={'button'}
								loading={resendLoading}
								disabled={loading}
							/>
						</div>
					</div>
				</form>
			</div>
		</div>
	)
}

