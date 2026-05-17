<template>
  <div class="card p-4 flex flex-wrap gap-3 items-end">
    <!-- Zeitraum -->
    <div class="flex flex-col gap-1">
      <label class="text-xs text-gray-400">Von</label>
      <input
        type="datetime-local"
        class="input"
        :value="modelValue.from"
        @input="update('from', ($event.target as HTMLInputElement).value)"
      />
    </div>
    <div class="flex flex-col gap-1">
      <label class="text-xs text-gray-400">Bis</label>
      <input
        type="datetime-local"
        class="input"
        :value="modelValue.to"
        @input="update('to', ($event.target as HTMLInputElement).value)"
      />
    </div>

    <!-- Action Filter -->
    <div class="flex flex-col gap-1">
      <label class="text-xs text-gray-400">Action</label>
      <select
        class="input"
        :value="modelValue.action"
        @change="update('action', ($event.target as HTMLSelectElement).value)"
      >
        <option value="">All</option>
        <option value="block">Block</option>
        <option value="detect">Detect</option>
      </select>
    </div>

    <!-- IP Suche -->
    <div class="flex flex-col gap-1">
      <label class="text-xs text-gray-400">Client IP</label>
      <input
        type="text"
        class="input w-40"
        placeholder="z.B. 1.2.3.4"
        :value="modelValue.client_ip"
        @input="update('client_ip', ($event.target as HTMLInputElement).value)"
      />
    </div>

    <!-- Quick ranges -->
    <div class="flex gap-1 items-end">
      <button class="btn-ghost text-xs" @click="setRange(1)">1h</button>
      <button class="btn-ghost text-xs" @click="setRange(6)">6h</button>
      <button class="btn-ghost text-xs" @click="setRange(24)">24h</button>
      <button class="btn-ghost text-xs" @click="clearRange">All</button>
    </div>

    <!-- Apply -->
    <div class="ml-auto flex items-end">
      <button class="btn-primary" @click="$emit('apply')">Filter anwenden</button>
    </div>
  </div>
</template>

<script setup lang="ts">
interface Filters {
  from: string
  to: string
  action: string
  client_ip: string
}

const props = defineProps<{ modelValue: Filters }>()

const emit = defineEmits<{
  'update:modelValue': [value: Filters]
  apply: []
}>()

function update(key: keyof Filters, value: string) {
  emit('update:modelValue', { ...props.modelValue, [key]: value })
}

function setRange(hours: number) {
  const now = new Date()
  const from = new Date(now.getTime() - hours * 60 * 60 * 1000)
  emit('update:modelValue', {
    from: toLocal(from),
    to: toLocal(now),
    action: '',
    client_ip: '',
  })
  emit('apply')
}

function clearRange() {
  emit('update:modelValue', { from: '', to: '', action: '', client_ip: '' })
  emit('apply')
}

function toLocal(d: Date): string {
  const offset = d.getTimezoneOffset()
  const local = new Date(d.getTime() - offset * 60 * 1000)
  return local.toISOString().slice(0, 16)
}
</script>
