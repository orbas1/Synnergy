;; log.wat - emits a log message via host interface
(module
  (import "env" "host_log" (func $host_log (param i32 i32)))
  (memory (export "memory") 1)
  (data (i32.const 0) "hello")
  (func (export "_start")
    (call $host_log (i32.const 0) (i32.const 5))
  )
)
