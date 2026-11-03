#!/bin/bash

# Writer 性能对比测试脚本（原始 vs Ultra）

echo "=================================="
echo "Writer 性能对比测试"
echo "原始版本 vs Ultra 优化版本"
echo "=================================="
echo ""

# 颜色定义
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# 测试配置
BENCHTIME="3s"
COUNT=3

echo -e "${BLUE}测试配置:${NC}"
echo "  - 每个测试运行时间: $BENCHTIME"
echo "  - 重复次数: $COUNT"
echo ""

# 1. Console Writer 对比
echo -e "${GREEN}[1/6] Console Writer 性能对比${NC}"
echo "----------------------------------------"
go test -bench='^BenchmarkConsoleWriter_Sequential$' -benchmem -benchtime=$BENCHTIME -count=$COUNT 2>/dev/null | grep "Benchmark"

# 2. Console Writer 并发对比
echo -e "${GREEN}[2/6] Console Writer 并发性能对比${NC}"
echo "----------------------------------------"
go test -bench='^BenchmarkConsoleWriter_Parallel$' -benchmem -benchtime=$BENCHTIME -count=$COUNT 2>/dev/null | grep "Benchmark"

# 3. File Writer 对比
echo -e "${GREEN}[3/6] File Writer 性能对比${NC}"
echo "----------------------------------------"
go test -bench='^BenchmarkFileWriter_Sequential$' -benchmem -benchtime=$BENCHTIME -count=$COUNT 2>/dev/null | grep "Benchmark"

# 4. File Writer 并发对比
echo -e "${GREEN}[4/6] File Writer 并发性能对比${NC}"
echo "----------------------------------------"
go test -bench='^BenchmarkFileWriter_Parallel$' -benchmem -benchtime=$BENCHTIME -count=$COUNT 2>/dev/null | grep "Benchmark"

# 5. Rotate Writer 对比
echo -e "${GREEN}[5/6] Rotate Writer 性能对比${NC}"
echo "----------------------------------------"
go test -bench='^BenchmarkRotateWriter_Sequential$' -benchmem -benchtime=$BENCHTIME -count=$COUNT 2>/dev/null | grep "Benchmark"

# 6. 统计信息对比
echo -e "${GREEN}[6/6] 统计信息性能对比${NC}"
echo "----------------------------------------"
echo "（顺序）:"
go test -bench='^BenchmarkStats_Original$' -benchmem -benchtime=$BENCHTIME -count=$COUNT 2>/dev/null | grep "Benchmark"
echo "（并发）:"
go test -bench='^BenchmarkStats_Original_Parallel$' -benchmem -benchtime=$BENCHTIME -count=$COUNT 2>/dev/null | grep "Benchmark"