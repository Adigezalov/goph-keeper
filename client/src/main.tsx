import { type Root, createRoot } from 'react-dom/client'

import { App } from '@app/app.tsx'

const container = document.getElementById('root') as Element
const root: Root = createRoot(container)

root.render(<App />)
