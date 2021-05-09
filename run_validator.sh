go build next_block.go
ln -s next_block next_block.exe
go build block_validator.go
echo "While the validator is running feed it with the test blocks by executing (in another terminal): ./run_java.sh"
./block_validator
rm block_validator next_block next_block.exe

