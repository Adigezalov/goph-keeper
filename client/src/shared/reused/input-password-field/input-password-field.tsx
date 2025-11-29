import { useMemo } from 'react'
import { Control, FieldPath, FieldValues, RegisterOptions } from 'react-hook-form'
import { useTranslation } from 'react-i18next'

import { InputPassword } from '@shared/uikit/input-password'

import { FormField } from '../form-field'

interface InputPasswordFieldProps<
	TFieldValues extends FieldValues = FieldValues,
	TName extends FieldPath<TFieldValues> = FieldPath<TFieldValues>,
> {
	control: Control<TFieldValues>
	name: TName
	label?: string
	placeholder?: string
	required?: boolean
	disabled?: boolean
	rules?: RegisterOptions<TFieldValues, TName>
	onBlur?: (value: string) => void
}

export const InputPasswordField = <
	TFieldValues extends FieldValues = FieldValues,
	TName extends FieldPath<TFieldValues> = FieldPath<TFieldValues>,
>({
	control,
	name,
	label,
	placeholder,
	required = false,
	disabled = false,
	rules,
	onBlur: customOnBlur,
}: InputPasswordFieldProps<TFieldValues, TName>) => {
	const { t } = useTranslation()

	const defaultRules = useMemo(() => {
		return required ? { required: t('validation_error.required_field'), ...rules } : rules
	}, [required, rules, t])

	return (
		<FormField control={control} name={name} rules={defaultRules}>
			{({ value, onChange, onBlur, invalid, error }) => (
				<InputPassword
					value={value || ''}
					label={label}
					placeholder={placeholder}
					required={required}
					invalid={invalid}
					disabled={disabled}
					error={error}
					onChange={onChange}
					onBlur={() => {
						const trimmedValue = value?.trim() || ''
						if (trimmedValue !== value) {
							onChange(trimmedValue)
						}
						if (customOnBlur) {
							customOnBlur(trimmedValue)
						}
						onBlur()
					}}
				/>
			)}
		</FormField>
	)
}
