<template>
  <div class="bg-gray-900 border border-gray-800 rounded-xl p-5">
    <h2 class="text-sm font-semibold text-gray-400 uppercase tracking-wider mb-4">Paranoia Level</h2>

    <!-- PL Enable/Disable Toggle -->
    <div class="flex items-center gap-3 mb-4">
      <button
        @click="$emit('update:enabled', !enabled)"
        :class="[
          'relative inline-flex h-6 w-11 flex-shrink-0 rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none',
          enabled ? 'bg-blue-600' : 'bg-gray-600'
        ]"
      >
        <span
          :class="[
            'pointer-events-none inline-block h-5 w-5 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out',
            enabled ? 'translate-x-5' : 'translate-x-0'
          ]"
        ></span>
      </button>
      <span class="text-sm text-gray-300">Paranoia Level aktivieren</span>
    </div>

    <!-- Disabled info box -->
    <div v-if="!enabled" class="rounded-lg border border-blue-500/30 bg-blue-950/20 p-4 text-xs text-blue-300">
      <div class="flex items-start gap-2">
        <svg class="w-4 h-4 flex-shrink-0 mt-0.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
        <span>Paranoia Level deaktiviert — Rules können unten manuell konfiguriert werden.</span>
      </div>
    </div>

    <!-- Slider content (shown only when enabled) -->
    <div v-show="enabled">
      <div class="flex gap-2 mb-4">
        <button
          v-for="lvl in [1,2,3,4]"
          :key="lvl"
          @click="$emit('update:modelValue', lvl)"
          :class="[
            'w-10 h-10 rounded-lg font-bold text-sm transition-all border',
            modelValue === lvl ? 'bg-blue-600 text-white border-blue-500' : 'bg-gray-800 text-gray-400 border-gray-700 hover:bg-gray-700'
          ]"
        >{{ lvl }}</button>
      </div>
      <div class="flex justify-between text-xs text-gray-500 mb-1">
        <span>Conservative</span>
        <span>Strict</span>
      </div>
      <div class="w-full h-1.5 bg-gray-700 rounded-full">
        <div class="h-1.5 bg-blue-500 rounded-full transition-all" :style="{ width: ((modelValue - 1) / 3 * 100) + '%' }"></div>
      </div>

      <Transition name="fade">
        <div
          :key="modelValue"
          class="mt-4 rounded-lg border p-4 text-xs"
          :class="infoBoxClasses[modelValue]"
        >
          <div class="flex items-center justify-between mb-2">
            <span class="font-semibold text-sm" :class="accentClasses[modelValue]">
              {{ PARANOIA_INFO[modelValue].label }}
            </span>
            <span class="text-gray-400 font-mono">{{ PARANOIA_INFO[modelValue].rules }}</span>
          </div>
          <p class="text-gray-300 mb-3 leading-relaxed">{{ PARANOIA_INFO[modelValue].description }}</p>
          <div class="mb-2">
            <span class="text-gray-500 uppercase tracking-wide text-[10px] font-semibold">Abgedeckte Kategorien</span>
            <div class="flex flex-wrap gap-1 mt-1">
              <span
                v-for="cat in PARANOIA_INFO[modelValue].categories"
                :key="cat"
                class="px-2 py-0.5 rounded bg-gray-800 text-gray-300"
              >{{ cat }}</span>
            </div>
          </div>
          <div class="flex items-center justify-between mt-3 pt-2 border-t border-gray-700/50">
            <span class="text-gray-400">
              <span class="text-gray-500">False Positives:</span>
              <span :class="accentClasses[modelValue]"> {{ PARANOIA_INFO[modelValue].falsePositives }}</span>
            </span>
          </div>
          <div class="mt-2 flex items-start gap-1 text-gray-400">
            <svg class="w-3.5 h-3.5 flex-shrink-0 mt-0.5 text-yellow-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z" />
            </svg>
            <span>{{ PARANOIA_INFO[modelValue].recommended }}</span>
          </div>
        </div>
      </Transition>
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{ modelValue: number; enabled: boolean }>()
defineEmits<{
  (e: 'update:modelValue', v: number): void
  (e: 'update:enabled', v: boolean): void
}>()

const PARANOIA_INFO: Record<number, {
  label: string
  rules: string
  description: string
  categories: string[]
  falsePositives: string
  recommended: string
}> = {
  1: {
    label: 'Level 1 — Standard',
    rules: '213 Rules aktiv (CRS v4)',
    description: 'Grundschutz gegen die häufigsten Angriffe. Kaum False Positives. Für die meisten Webapplikationen geeignet.',
    categories: ['SQL Injection (Basis)', 'XSS (Basis)', 'Path Traversal', 'Scanner-Erkennung'],
    falsePositives: 'Sehr gering',
    recommended: 'Empfohlen für Produktivumgebungen',
  },
  2: {
    label: 'Level 2 — Erhöht',
    rules: '302 Rules aktiv (CRS v4)',
    description: 'Erweiterte Erkennung mit aggressiverem Pattern-Matching. Erste False Positives bei komplexen Applikationen möglich.',
    categories: ['Alles aus PL1', 'SQLi (erweitert)', 'XSS (erweitert)', 'HTTP Protocol Attacks', 'Remote File Inclusion'],
    falsePositives: 'Gering bis moderat',
    recommended: 'Gut für sicherheitskritische Apps mit Tuning-Aufwand',
  },
  3: {
    label: 'Level 3 — Streng',
    rules: '332 Rules aktiv (CRS v4)',
    description: 'Sehr restriktiv. Viele False Positives — erfordert intensives Whitelisting. Nur mit sorgfältigem Tuning produktiv einsetzbar.',
    categories: ['Alles aus PL2', 'PHP Injection', 'Node.js Injection', 'Session Fixation', 'HTTP Splitting'],
    falsePositives: 'Hoch — Tuning erforderlich',
    recommended: 'Nur für High-Security mit Whitelist-Betrieb',
  },
  4: {
    label: 'Level 4 — Paranoid',
    rules: '341 Rules aktiv (CRS v4)',
    description: 'Maximale Erkennung. Extrem viele False Positives. Fast jeder komplexe Request triggert Rules. Nur für Hochsicherheitsumgebungen.',
    categories: ['Alles aus PL3', 'Erweiterte Encoding-Angriffe', 'Ultra-aggressive Pattern-Matches'],
    falsePositives: 'Sehr hoch — kaum produktiv nutzbar ohne umfangreiches Tuning',
    recommended: 'Nur für isolierte Hochsicherheitsumgebungen',
  },
}

const infoBoxClasses: Record<number, string> = {
  1: 'border-green-500/30 bg-green-950/20',
  2: 'border-yellow-500/30 bg-yellow-950/20',
  3: 'border-orange-500/30 bg-orange-950/20',
  4: 'border-red-500/30 bg-red-950/20',
}

const accentClasses: Record<number, string> = {
  1: 'text-green-400',
  2: 'text-yellow-400',
  3: 'text-orange-400',
  4: 'text-red-400',
}

const levelDot: Record<number, string> = {
  1: '🟢',
  2: '🟡',
  3: '🟠',
  4: '🔴',
}
</script>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.25s ease, transform 0.25s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}
</style>
