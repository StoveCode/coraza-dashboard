<template>
  <div class="min-h-screen bg-gray-950">
    <nav class="bg-gray-900 border-b border-gray-800 px-6 py-3 flex items-center gap-6">
      <span class="text-lg font-bold text-red-400">Coraza WAF</span>
      <router-link to="/" class="text-gray-300 hover:text-white transition-colors">Dashboard</router-link>
      <router-link to="/events" class="text-gray-300 hover:text-white transition-colors">Events</router-link>
      <router-link to="/rules" class="text-gray-300 hover:text-white transition-colors">Rules</router-link>
      <router-link to="/logs" class="text-gray-300 hover:text-white transition-colors">Logs</router-link>
    </nav>
    <router-view />
  </div>
</template>
