# <span style="color:#2b6cb0;">Data Aggregation Service</span>

<p>Сервис для обработки входящих пакетов данных, агрегации максимальных значений по временным периодам с доступом через HTTP REST и gRPC API.</p>

<hr>

<h2>Содержание</h2>
<ul>
  <li><a href="#сборка-и-запуск">Сборка и запуск</a></li>
  <li><a href="#использование">Использование</a></li>
  <li><a href="#рекомендации-для-продакшн">Рекомендации для продакшн</a></li>
  <li><a href="#мониторинг">Мониторинг</a></li>
</ul>

<hr>

<h2 id="сборка-и-запуск">Сборка и запуск</h2>

<h3>Требования</h3>
<ul>
  <li>Go 1.24+</li>
  <li>Docker и Docker Compose</li>
  <li>PostgreSQL 13+</li>
  <li>Protobuf compiler (<code>protoc</code>) с плагинами для Go</li>
</ul>

<h3>Локальная разработка с Docker Compose</h3>
<ol>
  <li>Создайте файл <code>.env</code> в корне проекта с параметрами подключения к БД и другими переменными окружения:
    <pre><code>PGUSER=youruser
PGPASSWORD=yourpassword
PGDATABASE=yourdb
PGHOST=localhost
PGPORT=5432
PGSSLMODE=disable
</code></pre>
  </li>
  <li>Запустите миграции и сервис командой:
    <pre><code>docker-compose up --build</code></pre>
  </li>
  <li>Сервис будет доступен по HTTP на порту <code>8080</code> и gRPC на порту <code>9090</code>.</li>
</ol>

<h3>Сборка и запуск вручную</h3>
<ol>
  <li>Сгенерируйте protobuf код:
    <pre><code>make proto</code></pre>
  </li>
  <li>Соберите бинарник:
    <pre><code>make build</code></pre>
  </li>
  <li>Запустите миграции (необходимо настроить переменные окружения):
    <pre><code>make migrate</code></pre>
  </li>
  <li>Запустите сервис:
    <pre><code>./data-aggregation-service</code></pre>
  </li>
</ol>

<hr>

<h2 id="использование">Использование</h2>

<h3>HTTP API</h3>
<ul>
  <li><code>GET /health</code> — проверка состояния сервиса</li>
  <li><code>GET /api/v1/max-values?start=&lt;RFC3339&gt;&amp;end=&lt;RFC3339&gt;</code> — получить максимальные значения за период</li>
  <li><code>GET /api/v1/max-values/{id}</code> — получить максимальное значение по ID пакета</li>
</ul>

<h3>gRPC API</h3>
<ul>
  <li><code>GetMaxValuesByPeriod(TimePeriod)</code> — получить максимальные значения за период</li>
  <li><code>GetMaxValueByID(PackageID)</code> — получить максимальное значение по ID пакета</li>
</ul>

<p>Описание protobuf в <code>api/proto/aggregator/v1/aggregator.proto</code>.</p>

<hr>

<h2 id="рекомендации-для-продакшн">Рекомендации для продакшн</h2>

<ol>
  <li><strong>Безопасность и секреты</strong>
    <ul>
      <li>Не хранить чувствительные данные (пароли, строки подключения) в Dockerfile или репозитории.</li>
      <li>Использовать секреты Kubernetes, HashiCorp Vault или аналогичные решения для управления конфигурацией.</li>
      <li>В Docker Compose использовать <code>.env</code> только для локальной разработки.</li>
    </ul>
  </li>
  <li><strong>Конфигурация</strong>
    <ul>
      <li>Использовать переменные окружения для настройки сервиса.</li>
      <li>Обеспечить возможность переопределения параметров без пересборки образа.</li>
    </ul>
  </li>
  <li><strong>База данных</strong>
    <ul>
      <li>Использовать пул соединений с оптимальными параметрами (максимальное число соединений, время жизни).</li>
      <li>Регулярно выполнять миграции схемы базы данных.</li>
      <li>Использовать партиционирование таблиц для масштабируемости (как в миграциях).</li>
    </ul>
  </li>
  <li><strong>Мониторинг и логирование</strong>
    <ul>
      <li>Включить сбор метрик Prometheus.</li>
      <li>Настроить централизованный сбор логов (Loki).</li>
      <li>Использовать уровни логирования (debug, info, warn, error).</li>
    </ul>
  </li>
  <li><strong>Отказоустойчивость</strong>
    <ul>
      <li>Настроить healthchecks и readiness probes.</li>
      <li>Обеспечить graceful shutdown для корректного завершения работы.</li>
      <li>Масштабировать количество воркеров под нагрузку.</li>
    </ul>
  </li>
  <li><strong>CI/CD</strong>
    <ul>
      <li>Автоматизировать сборку, тестирование, линтинг и деплой.</li>
      <li>Использовать мониторинг после деплоя.</li>
    </ul>
  </li>
</ol>

<hr>

<h2 id="мониторинг">Мониторинг</h2>

<p>Сервис экспортирует метрики в формате Prometheus на HTTP эндпоинте <code>/metrics</code> (порт 8080):</p>

<ul>
  <li>Количество HTTP-запросов (<code>http_requests_total</code>) с лейблами по методу, пути и статусу.</li>
  <li>Время обработки HTTP-запросов (<code>http_request_duration_seconds</code>) с лейблами по методу и пути.</li>
  <li>Размер HTTP-ответов в байтах (<code>http_response_size_bytes</code>).</li>
  <li>Количество gRPC-запросов (<code>grpc_requests_total</code>) с лейблами по методу и статусу.</li>
  <li>Время обработки gRPC-запросов (<code>grpc_request_duration_seconds</code>) с лейблами по методу и статусу.</li>
  <li>Время выполнения операций с базой данных (<code>db_query_duration_seconds</code>) с лейблом операции.</li>
  <li>Количество активных соединений с базой данных (<code>db_active_connections</code>).</li>
  <li>Количество простаивающих соединений с базой данных (<code>db_idle_connections</code>).</li>
  <li>Общее количество пакетов, полученных агрегатором (<code>aggregator_packets_received_total</code>).</li>
  <li>Количество успешно обработанных пакетов (<code>aggregator_packets_processed_total</code>).</li>
  <li>Количество пакетов, обработка которых завершилась ошибкой (<code>aggregator_packets_failed_total</code>).</li>
  <li>Гистограмма времени обработки пакета (<code>aggregator_packet_processing_seconds</code>).</li>
  <li>Текущее количество активных воркеров (<code>aggregator_active_workers</code>).</li>
</ul>


<h3>Пример настройки Prometheus для сбора метрик</h3>

<pre><code>scrape_configs:
  - job_name: 'data-aggregation-service'
    static_configs:
      - targets: ['app:8080']
    metrics_path: /metrics
    scrape_interval: 5s
</code></pre>

<h3>Рекомендации по мониторингу</h3>

<ul>
  <li>Настроить алерты на высокую задержку обработки пакетов и ошибки.</li>
  <li>Следить за состоянием базы данных и пулом соединений.</li>
  <li>Анализировать логи ошибок и предупреждений.</li>
  <li>Использовать Grafana для визуализации метрик.</li>
</ul>
