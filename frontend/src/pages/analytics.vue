<template>
  <div class="analytics">
    <h2 class="title">Аналитика</h2>
    <p v-if="link" class="subtitle">{{ link.original_url }}</p>

    <div v-if="loading" class="state muted">Загрузка...</div>
    <p v-else-if="error" class="state error">{{ error }}</p>

    <div v-else class="analytics-container">
      <!-- Легенда и информация -->
      <div class="info-section">
        <div class="stats">
          <div class="stat-card">
            <div class="stat-label">Всего переходов</div>
            <div class="stat-value">{{ transitions.length }}</div>
          </div>
          <div class="stat-card">
            <div class="stat-label">Уникальные города</div>
            <div class="stat-value">{{ uniqueCities }}</div>
          </div>
        </div>
        
        <div v-if="cityDistribution.length > 0" class="chart-container">
          <h3 class="chart-title">Распределение по городам</h3>
          <div class="chart-wrapper">
            <svg class="pie-chart" viewBox="0 0 200 200">
              <!-- Фон круга -->
              <circle cx="100" cy="100" r="80" fill="var(--bg-secondary)" />
              
              <!-- Секторы диаграммы -->
              <path 
                v-for="(item, index) in cityDistribution" 
                :key="item.city"
                :d="getArcPath(item.startAngle, item.endAngle)"
                :fill="getCityColor(index)"
                stroke="var(--bg-secondary)"
                stroke-width="2"
              />
              
              <!-- Центральный круг -->
              <circle cx="100" cy="100" r="30" fill="var(--bg-tertiary)" />
            </svg>
            
            <!-- Легенда -->
            <div class="legend">
              <div 
                v-for="(item, index) in cityDistribution" 
                :key="item.city"
                class="legend-item"
                @mouseenter="highlightSegment(index)"
                @mouseleave="unhighlightSegment(index)"
                :class="{ 'highlighted': highlightedIndex === index }"
              >
                <span class="legend-color" :style="{ backgroundColor: getCityColor(index) }"></span>
                <span class="legend-text">
                  <span class="city-name">{{ item.city || 'Не указан' }}</span>
                  <span class="city-count">({{ item.count }})</span>
                  <span class="city-percent">{{ item.percentage }}%</span>
                </span>
              </div>
            </div>
          </div>
        </div>
        <div v-else class="no-data">
          <p class="muted">Нет данных по городам для отображения диаграммы</p>
        </div>
      </div>

      <!-- Список переходов -->
      <div class="transitions-section">
        <div class="transitions-header">
          <div class="label">Переходы:</div>
          <div class="total">Всего: <strong>{{ transitions.length }}</strong></div>
        </div>
        
        <!-- Список переходов -->
        <div class="list">
          <div v-for="t in paginatedTransitions" :key="t.id" class="item">
            <div class="main">
              <div class="text">{{ formatRow(t) }}</div>
              <button class="ghost" @click="toggle(t)">{{ t._open ? 'Скрыть' : 'Подробнее' }}</button>
            </div>
            <div v-if="t._open" class="details">
              <div>Откуда: <strong>{{ t.referer || '—' }}</strong></div>
              <div>Браузер: <strong>{{ t.browser || '—' }}</strong></div>
              <div>OS: <strong>{{ t.os || '—' }}</strong></div>
              <div>User-agent: <strong class="ua">{{ t.user_agent || '—' }}</strong></div>
            </div>
          </div>
          <div v-if="transitions.length === 0" class="muted dots">Нет переходов</div>
        </div>
        
        <!-- Пагинация -->
        <div v-if="totalPages > 1" class="pagination">
          <button 
            class="pagination-btn" 
            @click="prevPage" 
            :disabled="currentPage === 1"
          >
            ←
          </button>
          
          <div class="page-numbers">
            <span v-if="totalPages <= 7">
              <button 
                v-for="page in totalPages" 
                :key="page"
                class="page-btn"
                :class="{ active: currentPage === page }"
                @click="goToPage(page)"
              >
                {{ page }}
              </button>
            </span>
            <span v-else>
              <button 
                v-if="currentPage > 3"
                class="page-btn"
                @click="goToPage(1)"
              >
                1
              </button>
              <span v-if="currentPage > 4" class="ellipsis">...</span>
              
              <button 
                v-for="page in getPageRange" 
                :key="page"
                class="page-btn"
                :class="{ active: currentPage === page }"
                @click="goToPage(page)"
              >
                {{ page }}
              </button>
              
              <span v-if="currentPage < totalPages - 3" class="ellipsis">...</span>
              <button 
                v-if="currentPage < totalPages - 2"
                class="page-btn"
                @click="goToPage(totalPages)"
              >
                {{ totalPages }}
              </button>
            </span>
          </div>
          
          <button 
            class="pagination-btn" 
            @click="nextPage" 
            :disabled="currentPage === totalPages"
          >
            →
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { onMounted, ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'

const route = useRoute()
const router = useRouter()
const id = Number(route.params.id)

const loading = ref(false)
const error = ref(null)
const link = ref(null)
const transitions = ref([])
const currentPage = ref(1)
const itemsPerPage = 10
const highlightedIndex = ref(-1)

// Цветовая палитра для диаграммы
const cityColors = [
  'var(--chart-color-1)',
  'var(--chart-color-2)',
  'var(--chart-color-3)',
  'var(--chart-color-4)',
  'var(--chart-color-5)',
  'var(--chart-color-6)',
  'var(--chart-color-7)',
  'var(--chart-color-8)'
]

function formatRow(t) {
  const country = t.country || 'Unknown'
  const city = t.city || ''
  const dt = new Date(t.created_at)
  const dateStr = dt.toLocaleString()
  return `${country}${city ? ', ' + city : ''} ${dateStr}`
}

function toggle(t) { t._open = !t._open }

// Пагинация
const totalPages = computed(() => Math.ceil(transitions.value.length / itemsPerPage))
const paginatedTransitions = computed(() => {
  const start = (currentPage.value - 1) * itemsPerPage
  const end = start + itemsPerPage
  return transitions.value.slice(start, end)
})

const getPageRange = computed(() => {
  const range = []
  let start = Math.max(1, currentPage.value - 2)
  let end = Math.min(totalPages.value, currentPage.value + 2)
  
  if (currentPage.value <= 3) {
    end = Math.min(5, totalPages.value)
  }
  if (currentPage.value >= totalPages.value - 2) {
    start = Math.max(totalPages.value - 4, 1)
  }
  
  for (let i = start; i <= end; i++) {
    range.push(i)
  }
  return range
})

function goToPage(page) {
  currentPage.value = page
  window.scrollTo({ top: document.querySelector('.transitions-section').offsetTop, behavior: 'smooth' })
}

function prevPage() {
  if (currentPage.value > 1) {
    currentPage.value--
    window.scrollTo({ top: document.querySelector('.transitions-section').offsetTop, behavior: 'smooth' })
  }
}

function nextPage() {
  if (currentPage.value < totalPages.value) {
    currentPage.value++
    window.scrollTo({ top: document.querySelector('.transitions-section').offsetTop, behavior: 'smooth' })
  }
}

// Распределение по городам
const uniqueCities = computed(() => {
  const cities = new Set()
  transitions.value.forEach(t => {
    if (t.city) cities.add(t.city)
  })
  return cities.size
})

const cityDistribution = computed(() => {
  const cityCounts = {}
  transitions.value.forEach(t => {
    const city = t.city || 'Не указан'
    cityCounts[city] = (cityCounts[city] || 0) + 1
  })
  
  const total = transitions.value.length
  if (total === 0) return []
  
  // Сортируем по убыванию количества
  const sorted = Object.entries(cityCounts)
    .map(([city, count]) => ({ city, count }))
    .sort((a, b) => b.count - a.count)
  
  // Ограничиваем количество отображаемых городов (первые 7)
  const topCities = sorted.slice(0, 7)
  
  // Объединяем остальные в "Другие"
  if (sorted.length > 7) {
    const otherCount = sorted.slice(7).reduce((sum, item) => sum + item.count, 0)
    if (otherCount > 0) {
      topCities.push({ city: 'Другие', count: otherCount })
    }
  }
  
  // Рассчитываем углы для диаграммы
  let currentAngle = 0
  return topCities.map(item => {
    const angle = (item.count / total) * 360
    const startAngle = currentAngle
    const endAngle = currentAngle + angle
    const percentage = ((item.count / total) * 100).toFixed(1)
    
    currentAngle += angle
    
    return {
      ...item,
      percentage,
      startAngle,
      endAngle
    }
  })
})

// Функции для диаграммы
function getArcPath(startAngle, endAngle) {
  const radius = 80
  const centerX = 100
  const centerY = 100
  
  const startRad = (startAngle - 90) * Math.PI / 180
  const endRad = (endAngle - 90) * Math.PI / 180
  
  const startX = centerX + radius * Math.cos(startRad)
  const startY = centerY + radius * Math.sin(startRad)
  
  const endX = centerX + radius * Math.cos(endRad)
  const endY = centerY + radius * Math.sin(endRad)
  
  const largeArcFlag = (endAngle - startAngle) > 180 ? 1 : 0
  
  return [
    `M ${centerX} ${centerY}`,
    `L ${startX} ${startY}`,
    `A ${radius} ${radius} 0 ${largeArcFlag} 1 ${endX} ${endY}`,
    `L ${centerX} ${centerY}`
  ].join(' ')
}

function getCityColor(index) {
  return cityColors[index % cityColors.length]
}

function highlightSegment(index) {
  highlightedIndex.value = index
}

function unhighlightSegment(index) {
  highlightedIndex.value = -1
}

async function fetchData() {
  loading.value = true
  error.value = null
  try {
    const [linkRes, tranRes] = await Promise.all([
      fetch(`/api/v1/links/${id}`, { credentials: 'include' }),
      fetch(`/api/v1/links/${id}/transitions`, { credentials: 'include' }),
    ])
    if (!linkRes.ok) {
      if (linkRes.status === 401) { router.push('/login'); return }
      throw new Error('Не удалось получить ссылку')
    }
    if (!tranRes.ok) {
      if (tranRes.status === 401) { router.push('/login'); return }
      throw new Error('Не удалось получить аналитику')
    }
    link.value = await linkRes.json()
    const data = await tranRes.json()
    transitions.value = Array.isArray(data?.transitions) ? data.transitions : []
  } catch (e) {
    error.value = e.message || 'Ошибка'
  } finally {
    loading.value = false
  }
}

onMounted(fetchData)
</script>

<style scoped>
.analytics { 
  max-width: 1200px; 
  margin: 60px auto; 
  padding: 0 16px; 
}

.title { 
  margin: 0; 
  color: var(--text-primary); 
  font-weight: 600; 
  font-size: 28px;
}

.subtitle { 
  color: var(--text-secondary); 
  margin: 8px 0 24px; 
  font-size: 16px;
}

.state { margin: 24px 0; }
.muted { color: var(--text-muted); }
.error { color: var(--error-color); }

.analytics-container {
  display: grid;
  grid-template-columns: 1fr 1.5fr;
  gap: 30px;
  margin-top: 30px;
}

@media (max-width: 1024px) {
  .analytics-container {
    grid-template-columns: 1fr;
    gap: 40px;
  }
}

/* Секция с диаграммой */
.info-section {
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: 12px;
  padding: 24px;
  height: fit-content;
  box-shadow: 0 2px 8px var(--shadow-light);
}

.stats {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
  margin-bottom: 24px;
}

.stat-card {
  background: var(--bg-tertiary);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 16px;
  text-align: center;
}

.stat-label {
  color: var(--text-secondary);
  font-size: 14px;
  margin-bottom: 8px;
}

.stat-value {
  color: var(--text-primary);
  font-size: 24px;
  font-weight: 600;
}

.chart-title {
  color: var(--text-primary);
  font-size: 18px;
  font-weight: 600;
  margin: 0 0 20px 0;
  text-align: center;
}

.chart-wrapper {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 24px;
}

.pie-chart {
  width: 200px;
  height: 200px;
}

.legend {
  width: 100%;
}

.legend-item {
  display: flex;
  align-items: center;
  padding: 8px 12px;
  margin-bottom: 6px;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s ease;
  border: 1px solid transparent;
}

.legend-item:hover {
  background: var(--bg-tertiary);
  border-color: var(--border-color);
}

.legend-item.highlighted {
  background: var(--bg-tertiary);
  border-color: var(--accent-color);
  box-shadow: 0 0 0 1px var(--accent-color);
}

.legend-color {
  width: 12px;
  height: 12px;
  border-radius: 50%;
  margin-right: 12px;
  flex-shrink: 0;
}

.legend-text {
  display: flex;
  align-items: center;
  gap: 8px;
  color: var(--text-primary);
  font-size: 14px;
  width: 100%;
}

.city-name {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.city-count {
  color: var(--text-secondary);
  font-weight: 500;
}

.city-percent {
  color: var(--accent-color);
  font-weight: 600;
  margin-left: auto;
}

.no-data {
  text-align: center;
  padding: 40px 0;
  color: var(--text-muted);
}

/* Секция со списком переходов */
.transitions-section {
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: 12px;
  padding: 24px;
  box-shadow: 0 2px 8px var(--shadow-light);
}

.transitions-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  padding-bottom: 16px;
  border-bottom: 1px solid var(--border-color);
}

.label { 
  font-weight: 600; 
  color: var(--text-primary); 
  font-size: 18px;
}

.total { 
  color: var(--text-primary); 
  font-size: 16px;
}

.list { 
  display: grid; 
  gap: 12px; 
  margin-bottom: 24px;
  min-height: 500px;
}

.item { 
  background: var(--bg-tertiary); 
  border: 1px solid var(--border-color); 
  border-radius: 10px; 
  padding: 16px; 
  transition: all 0.2s ease;
}

.item:hover {
  border-color: var(--border-hover);
  box-shadow: 0 2px 8px var(--shadow-light);
}

.main { 
  display: flex; 
  justify-content: space-between; 
  align-items: center; 
  gap: 12px; 
}

.text { 
  font-weight: 600; 
  color: var(--text-primary); 
  font-size: 14px;
  flex: 1;
  min-width: 0;
}

.ghost { 
  background: var(--bg-secondary); 
  color: var(--text-primary); 
  border: 1px solid var(--border-color); 
  border-radius: 8px; 
  padding: 8px 16px; 
  cursor: pointer; 
  font-size: 14px;
  transition: all 0.2s ease;
  white-space: nowrap;
  flex-shrink: 0;
}

.ghost:hover { 
  background: var(--bg-tertiary); 
  transform: translateY(-1px); 
  box-shadow: 0 2px 8px var(--shadow-light);
}

.details { 
  margin-top: 16px; 
  padding: 16px; 
  border-left: 4px solid var(--accent-color); 
  background: var(--bg-secondary); 
  border-radius: 8px; 
  color: var(--text-primary); 
  font-size: 14px;
}

.details > div {
  margin-bottom: 8px;
}

.details > div:last-child {
  margin-bottom: 0;
}

.ua { 
  word-break: break-all; 
  font-size: 12px;
  color: var(--text-secondary);
}

.dots { 
  text-align: center; 
  padding: 40px 0;
}

/* Пагинация */
.pagination {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 16px;
  padding: 16px;
  border-top: 1px solid var(--border-color);
  margin-top: 20px;
}

.pagination-btn {
  width: 40px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg-tertiary);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  color: var(--text-primary);
  cursor: pointer;
  transition: all 0.2s ease;
  font-size: 18px;
}

.pagination-btn:hover:not(:disabled) {
  background: var(--bg-secondary);
  border-color: var(--border-hover);
  transform: translateY(-1px);
  box-shadow: 0 2px 8px var(--shadow-light);
}

.pagination-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.page-numbers {
  display: flex;
  align-items: center;
  gap: 8px;
}

.page-btn {
  width: 40px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg-tertiary);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  color: var(--text-primary);
  cursor: pointer;
  transition: all 0.2s ease;
  font-size: 14px;
  font-weight: 500;
}

.page-btn:hover {
  background: var(--bg-secondary);
  border-color: var(--border-hover);
}

.page-btn.active {
  background: var(--accent-color);
  border-color: var(--accent-color);
  color: white;
}

.ellipsis {
  color: var(--text-secondary);
  padding: 0 4px;
}

/* Темная тема - дополнительные настройки */
:root.dark-theme .pie-chart circle {
  fill: var(--bg-tertiary);
}

:root.dark-theme .legend-item:hover {
  background: var(--bg-secondary);
}

:root.dark-theme .item:hover {
  background: var(--bg-secondary);
}
</style>

<style>

:root {
  --chart-color-1: #4CAF50;
  --chart-color-2: #2196F3;
  --chart-color-3: #FF9800;
  --chart-color-4: #F44336;
  --chart-color-5: #9C27B0;
  --chart-color-6: #00BCD4;
  --chart-color-7: #8BC34A;
  --chart-color-8: #FFC107;
}

:root.dark-theme {
  --chart-color-1: #66BB6A;
  --chart-color-2: #42A5F5;
  --chart-color-3: #FFA726;
  --chart-color-4: #EF5350;
  --chart-color-5: #AB47BC;
  --chart-color-6: #26C6DA;
  --chart-color-7: #9CCC65;
  --chart-color-8: #FFCA28;
}
</style>