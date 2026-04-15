<script setup lang="ts">
import { ref } from "vue";

const count = ref<number>(0);

const handleClick = () => {
  count.value++
}
</script>

<template>
  <main>
    <h1>Welcome!</h1>
    <div>
      <span>
        Count: {{ count }}
      </span>
      <button @click="handleClick">Click me</button>
    </div>
  </main>
</template>

