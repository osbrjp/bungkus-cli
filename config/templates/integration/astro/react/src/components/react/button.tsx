import { useState } from "react"

const Button = () => {
  const [count, setCount] = useState<number>(0)
  const handleClick = () => {
    setCount((c) => c + 1)
    console.log("Count: ", count)
  }
  return <>
    <span>
      Count: {count}
    </span>
    <button type="button" onClick={handleClick}>
      Click me
    </button>
  </>
}
export default Button
