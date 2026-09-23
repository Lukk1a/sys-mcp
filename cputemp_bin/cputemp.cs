using System;
using LibreHardwareMonitor.Hardware;

public class UpdateVisitor : IVisitor
{
    public void VisitComputer(IComputer computer)
    {
        computer.Traverse(this);
    }
    public void VisitHardware(IHardware hardware)
    {
        hardware.Update();
        foreach (IHardware subHardware in hardware.SubHardware) subHardware.Accept(this);
    }
    public void VisitSensor(ISensor sensor) { }
    public void VisitParameter(IParameter parameter) { }
}

class Program
{
    static void Main(string[] args)
    {
        try
        {
            Computer computer = new Computer
            {
                IsCpuEnabled = true,
                IsMotherboardEnabled = true
            };

            computer.Open();
            computer.Accept(new UpdateVisitor());

            bool found = false;
            
            Action<IHardware> checkHardware = null;
            checkHardware = (hw) =>
            {
                bool isCpuHardware = (hw.HardwareType == HardwareType.Cpu);
                foreach (ISensor sensor in hw.Sensors)
                {
                    if (sensor.SensorType == SensorType.Temperature)
                    {
                        if (isCpuHardware || sensor.Name.IndexOf("CPU", StringComparison.OrdinalIgnoreCase) >= 0 || sensor.Name.IndexOf("Core", StringComparison.OrdinalIgnoreCase) >= 0)
                        {
                            if (sensor.Value.HasValue)
                            {
                                Console.WriteLine(String.Format("{0} ({1}): {2:F2}°C", sensor.Name, hw.Name, sensor.Value.Value));
                                found = true;
                            }
                        }
                    }
                }
                foreach (IHardware sub in hw.SubHardware)
                {
                    checkHardware(sub);
                }
            };

            foreach (IHardware hardware in computer.Hardware)
            {
                checkHardware(hardware);
            }
            
            computer.Close();
            
            if (!found) {
                Console.WriteLine("No CPU temperature sensors found.");
            }
        }
        catch (Exception ex)
        {
            Console.WriteLine(String.Format("Error: {0}", ex.Message));
        }
    }
}
