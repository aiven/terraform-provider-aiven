resource "aiven_flink_job" "job" {                                                                  
    project = aiven_flink.flink.project                                                             
    service_name = aiven_flink.flink.service_name                                                   
    job_name = "<JOB_NAME>"                                                                         
                                                                                                    
    table_id = [                                                                                      
        aiven_flink_table.source.table_id,                                                          
        aiven_flink_table.sink.table_id,                                                            
    ]                                                                                               
                                                                                                    
    statement = <<EOF                                                                              
        INSERT INTO ${aiven_flink_table.sink.table_name}                                            
        SELECT * FROM ${aiven_flink_table.source.table_name}                                        
        WHERE `cpu` > 50                                                                            
    EOF                                                                                             
}                                                                                                   
