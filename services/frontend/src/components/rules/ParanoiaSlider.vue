<template>
  <div class="bg-gray-900 border border-gray-800 rounded-xl p-5">
    <h2 class="text-sm font-semibold text-gray-400 uppercase tracking-wider mb-4">Paranoia Level</h2>
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
            {{ levelDot[modelValue] }} {{ PARANOIA_INFO[modelValue].label }}
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
          <span>💡</span>
          <span>{{ PARANOIA_INFO[modelValue].recommended }}</span>
        </div>
      </div>
    </Transition>
  </div>
</template>

<script setup lang="ts">
defineProps<{ modelValue: number }>()
defineEmits<{ (e: 'update:modelValue', v: number): void }>()

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
    rules: '~500 Rules aktiv',
    description: 'Grundschutz gegen die häufigsten Angriffe. Kaum False Positives. Für die meisten Webapplikationen geeignet.',
    categories: ['SQL Injection (Basis)', 'XSS (Basis)', 'Path Traversal', 'Scanner-Erkennung'],
    falsePositives: 'Sehr gering',
    recommended: 'Empfohlen für Produktivumgebungen',
  },
  2: {
    label: 'Level 2 — Erhöht',
    rules: '~900 Rules aktiv',
    description: 'Erweiterte Erkennung mit aggressiverem Pattern-Matching. Erste False Positives bei komplexen Applikationen möglich.',
    categories: ['Alles aus PL1', 'SQLi (erweitert)', 'XSS (erweitert)', 'HTTP Protocol Attacks', 'Remote File Inclusion'],
    falsePositives: 'Gering bis moderat',
    recommended: 'Gut für sicherheitskritische Apps mit Tuning-Aufwand',
  },
  3: {
    label: 'Level 3 — Streng',
    rules: '~1200 Rules aktiv',
    description: 'Sehr restriktiv. Viele False Positives — erfordert intensives Whitelisting. Nur mit sorgfältigem Tuning produktiv einsetzbar.',
    categories: ['Alles aus PL2', 'PHP Injection', 'Node.js Injection', 'Session Fixation', 'HTTP Splitting'],
    falsePositives: 'Hoch — Tuning erforderlich',
    recommended: 'Nur für High-Security mit Whitelist-Betrieb',
  },
  4: {
    label: 'Level 4 — Paranoid',
    rules: '~1400 Rules aktiv',
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
