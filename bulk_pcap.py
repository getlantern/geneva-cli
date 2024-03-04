import os
import subprocess
import json
import sys

from pathlib import Path

def main():

    script_dir = Path(os.path.dirname(os.path.realpath(__file__)))


    if len(sys.argv) < 2:
        print('No input file given')
        exit(1)

    input_file = sys.argv[1]
    strat_file = script_dir / 'strategies.txt'
    output_dir = script_dir / 'testdata/pcap_output'
    log_file = script_dir / 'testdata/log.txt'
    map_file = output_dir / 'map_file.json'
    
    run_results = {'good':{}, 'bad':{}}
    
    print(f'Output folder: {output_dir}')
    print(f'Results summary: {map_file}')

    with open(strat_file, 'r') as read_f:
        
        temp = read_f.readlines()
        for i, strat in enumerate((_.strip() for _ in temp if _.strip())):            
            outfile = output_dir / f'output_{i}.pcap'
            
            geneva_exe = script_dir / 'geneva-cli.exe'
            try: 
                print(f'[{i}] Running {strat}')
                results = subprocess.run([geneva_exe, 'run-pcap', '-f', '-s', strat, '--o', outfile,'--i', input_file],
                                        capture_output=True, check=True)

                output = results.stdout.decode("utf-8")
                for line in output.split('\n'):
                    print(line)
                run_results['good'][i] = {'strategy': strat, 'output':results.stdout.decode("utf-8")}
            except subprocess.CalledProcessError as err:
                print(f'Error in strategy {i} \n\t {strat}')
                print(err.output)
                run_results['bad'][i] = {'strategy': strat, 'output': results.stderr.decode("utf-8")}
        
        with open(map_file, 'w') as write_f:
            write_f.write(json.dumps(run_results, indent=4, sort_keys=4))

        for key, value in run_results['bad'].items():
            print(f'{key}: {value}')

if __name__ == '__main__':
    main()