/// <reference types="vite/client" />

// Declare raw imports for markdown files
declare module '*.md?raw' {
  const content: string
  export default content
}

// Declare raw imports for any file
declare module '*?raw' {
  const content: string
  export default content
}
