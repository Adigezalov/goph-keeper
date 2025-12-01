import react from '@vitejs/plugin-react'
import * as path from 'path'
import { defineConfig } from 'vite'

export default defineConfig({
	server: {
		port: 3000,
		host: 'localhost',
		proxy: {
			'/api': {
				target: 'http://localhost:8080',
				changeOrigin: true,
				secure: false,
				ws: true,
				configure: (proxy, _options) => {
					proxy.on('error', (err, _req, _res) => {
						console.log('proxy error', err)
					})
					proxy.on('proxyReqWs', (_proxyReq, req, _socket) => {
						console.log('WebSocket proxy request:', req.url)
					})
					proxy.on('proxyReq', (_proxyReq, req, _res) => {
						console.log('HTTP proxy request:', req.method, req.url)
					})
				},
			},
		},
	},
	plugins: [react()],
	resolve: {
		alias: {
			'@app': path.resolve(__dirname, './src/app'),
			'@entities': path.resolve(__dirname, './src/entities'),
			'@features': path.resolve(__dirname, './src/features'),
			'@pages': path.resolve(__dirname, './src/pages'),
			'@shared': path.resolve(__dirname, './src/shared'),
			'@widgets': path.resolve(__dirname, './src/widgets'),
		},
		extensions: ['.ts', '.tsx', '.js', '.jsx'],
	},
	build: {
		outDir: './dist',
		emptyOutDir: true,
		reportCompressedSize: true,
		chunkSizeWarningLimit: 1000,
		commonjsOptions: {
			transformMixedEsModules: true,
		},
		rollupOptions: {
			external: ['quill'],
			output: {
				manualChunks: {
					vendor: ['react', 'react-dom'],
					router: ['react-router-dom'],
					ui: ['primereact'],
					state: ['mobx', 'mobx-react-lite'],
					utils: ['axios', 'dayjs', 'i18next'],
				},
				chunkFileNames: (chunkInfo) => {
					const facadeModuleId = chunkInfo.facadeModuleId
					if (facadeModuleId && facadeModuleId.includes('pages/')) {
						const pageName = facadeModuleId.split('pages/')[1].split('/')[0]
						return `pages/${pageName}-[hash].js`
					}
					return 'chunks/[name]-[hash].js'
				},
				entryFileNames: 'assets/[name]-[hash].js',
			},
		},
	},
	css: {
		modules: {
			localsConvention: 'camelCase',
			generateScopedName: '[name]__[local]___[hash:base64:5]',
		},
		preprocessorOptions: {
			scss: {
				quietDeps: true,
			},
			sass: {
				quietDeps: true,
			},
		},
	},
})
