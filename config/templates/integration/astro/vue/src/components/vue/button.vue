<script setup lang="ts">
import { ref } from "vue"

const count = ref<number>(0)

const handleClick = () => {
  count.value++
  console.log("Count: ", count.value)
}
</script>

<template>
  <div>
    <span>
      Count: {{ count }} 
    </span>

    <button @click="handleClick">Click me</button>
  </div>
</template>
