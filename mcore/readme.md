# 


tbot -> message -> mcore -> dialogMng - createDialog -> taskMng - createTask for issue  -> dialogMng - if needed Generate answer -> return tbot

taskRunner  get-task -> mcore - take task -> taskRunner Run task in background
task - try todo task - report todo or error -> mcore - get report -> taskMng get task and mark it -> dilogMng - get dialog 