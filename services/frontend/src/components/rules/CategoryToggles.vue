<template>
  <div class="bg-gray-900 border border-gray-800 rounded-xl p-5">
    <h2 class="text-sm font-semibold text-gray-400 uppercase tracking-wider mb-4">CRS Rule Categories</h2>
    <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
      <div
        v-for="cat in categories"
        :key="cat.tag"
        class="flex items-center justify-between bg-gray-800 rounded-lg px-4 py-3"
      >
        <div>
          <p class="text-sm text-gray-200 font-medium">{{ cat.label }}</p>
          <p class="text-xs text-gray-500 mt-0.5">{{ cat.description }}</p>
        </div>
        <button
          @click="$emit('toggle', cat.tag)"
          :class="[
            'relative inline-flex h-6 w-11 flex-shrink-0 rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none ml-4',
            isDisabled(cat.tag) ? 'bg-gray-600' : 'bg-green-500'
          ]"
          :title="isDisabled(cat.tag) ? 'Enable' : 'Disable'"
        >
          <span
            :class="[
              'pointer-events-none inline-block h-5 w-5 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out',
              isDisabled(cat.tag) ? 'translate-x-0' : 'translate-x-5'
            ]"
          ></span>
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { RuleCategory } from '../../api/rules'

const props = defineProps<{
  categories: RuleCategory[]
  disabledTags: string[]
}>()

defineEmits<{ (e: 'toggle', tag: string): void }>()

function isDisabled(tag: string): boolean {
  return props.disabledTags.includes(tag)
}
</script>
